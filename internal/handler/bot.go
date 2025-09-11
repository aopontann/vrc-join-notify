package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	disc "github.com/aopontann/vrc-join-notify/internal/discord"
	"github.com/aopontann/vrc-join-notify/internal/firestore"
	vrc2 "github.com/aopontann/vrc-join-notify/internal/vrc"
	"github.com/bwmarrin/discordgo"
)

func DiscordBotHandler(db *firestore.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		discord, _ := disc.New()

		// Discord インタラクションの事前認証
		ok, err := discord.VerifyInteraction(r)
		if err != nil {
			http.Error(w, "Error decoding hex string", http.StatusInternalServerError)
			return
		}
		if !ok {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		// リクエストボディ解析
		// イベントタイプごとにデータが違うため注意
		// ただ、タイプ種類は必ずデータに含まれているため、タイプ種類を取得する
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading request body: " + err.Error())
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		fmt.Println("Request body:", string(body))

		eventType, userID, channelID, interactionData, err := discord.GetEventInfo(body)

		// Webhooks登録時の処理（イベント発生時にリクエストを飛ばすURLを登録する際に必要な処理　初回のみ
		// Interactions Endpoint URL登録時に必要な処理
		if eventType == disc.WebhooksResist || eventType == disc.InteractionsEndpointResist {
			pongResp, err := json.Marshal(discordgo.InteractionResponse{
				Type: discordgo.InteractionResponsePong,
			})
			if err != nil {
				http.Error(w, "Error marshalling response", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(pongResp)
			if err != nil {
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// アプリをインストールしたときのイベント処理
		if eventType == disc.ApplicationInstall {
			// DMチャンネルを作成
			ch, err := discord.UserChannelCreate(userID)
			if err != nil {
				http.Error(w, "Error creating user channel", http.StatusInternalServerError)
				return
			}

			slog.Info("Creating user channel", "channel_id", ch.ID)

			// チャンネルにメッセージを送信
			c := `
アプリのインストールが完了しました。
指定したフレンドが「だれでもおいで」ステータスになったとき通知を受け取れるアプリです。
操作は基本的にスラッシュコマンドを用います。
スラッシュコマンドは画面下のゲームパッドボタンから実行できます。
スラッシュコマンド一覧
- /auth login		ログイン
- /auth logout		ログアウト
- /auth email-code	2段階認証
- /join register	通知登録

使い方
1. ユーザ名とパスワードを指定してログインコマンドを実行してください。
2. 「メールに認証コードが送信されました。」と表示された場合は、メールに届いた番号を指定して2段階認証コマンドを実行してください。
3. 通知を受け取りたいフレンドのVRChatのWEBページのURLを指定して、通知登録コマンドを実行してください。
4. 指定したフレンドが「だれでもおいで」ステータスになったとき通知が届きます。

備考
- 「再ログインしてください。」とメッセージが届いた場合、ログインコマンドを再実行してください。
- 通知対象のフレンドは1人のみ指定できます。
- 他のフレンドの通知を受け取りたい場合、通知登録コマンドを再実行してください。
- フレンドのVRChatのWEBページのURLは次のようなものです。（https://vrchat.com/home/user/XXXXXX）
`

			if _, err := discord.ChannelMessageSend(ch.ID, c); err != nil {
				http.Error(w, "Error sending message", http.StatusInternalServerError)
				return
			}
			slog.Info("Received interaction type 1, sending response")

			// チャンネルIDとユーザIDを保存する処理を追加
			err = db.SaveUserInfo(userID, ch.ID)
			if err != nil {
				http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
				return
			}
		}

		// スラッシュコマンドの実行時の処理
		if eventType == disc.SlashCommand {
			// 認証関連処理
			if interactionData.Name == "auth" {
				vrc := vrc2.NewVRC()
				subCmdInfo := interactionData.Options[0]

				// ログイン処理（ユーザ名とパスワードを取得）
				if subCmdInfo.Name == "login" {
					username := subCmdInfo.Options[0].Value
					password := subCmdInfo.Options[1].Value
					token, err := vrc.Login(username, password)
					if err != nil {
						if _, err := discord.ChannelMessageSend(channelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					err = db.SaveUserToken(userID, token)
					if err != nil {
						if _, err := discord.ChannelMessageSend(channelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					// トークンが有効かされたかチェック　メール認証が必要な場合は無効になる
					ok, err := vrc.VerifyAuthToken(token)
					if err != nil {
						ErrorHandler(w, err, http.StatusInternalServerError)
					}

					if !ok {
						if _, err := discord.ChannelMessageSend(channelID, "メールに認証コードが送信されました。"); err != nil {
							ErrorHandler(w, err, http.StatusInternalServerError)
						}
					} else {
						if err := db.ChangeNotificationed(userID, false); err != nil {
							ErrorHandler(w, err, http.StatusInternalServerError)
						}
						if _, err := discord.ChannelMessageSend(channelID, "ログインしました。"); err != nil {
							ErrorHandler(w, err, http.StatusInternalServerError)
						}
					}
				}

				if subCmdInfo.Name == "logout" {
					if _, err := discord.ChannelMessageSend(channelID, "未実装"); err != nil {
						http.Error(w, "Error sending message", http.StatusInternalServerError)
						return
					}
				}

				// ログイン処理（2FA）
				if subCmdInfo.Name == "email-code" {
					code := subCmdInfo.Options[0].Value
					// DBからユーザのトークンを取得
					userInfo, err := db.GetUserInfo(userID)
					if err != nil {
						if _, err := discord.ChannelMessageSend(channelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					twoFactorAuthToken, err := vrc.Verify2FA(code, userInfo.Token)
					if err != nil {
						if _, err := discord.ChannelMessageSend(channelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					err = db.SaveUserTwoFactorAuthToken(userID, twoFactorAuthToken)
					if err != nil {
						if _, err := discord.ChannelMessageSend(channelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					if _, err := discord.ChannelMessageSend(channelID, "OK"); err != nil {
						http.Error(w, "Error sending message", http.StatusInternalServerError)
						return
					}
				}
			}

			// JOIN通知関連処理
			if interactionData.Name == "join" {
				subCmdInfo := interactionData.Options[0]

				// JOIN通知対象のユーザIDを登録
				if subCmdInfo.Name == "register" {
					url := subCmdInfo.Options[0].Value // https://vrchat.com/home/user/usr_33a8da12-14f4-4225-8711-320471ceb60b
					targetUserID := url[len("https://vrchat.com/home/user/"):]
					err := db.SaveTargetUser(userID, targetUserID)
					if err != nil {
						if _, err := discord.ChannelMessageSend(channelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					} else {
						if _, err := discord.ChannelMessageSend(channelID, "OK"); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			http.Error(w, "Error sending message", http.StatusInternalServerError)
			return
		}
	}
}
