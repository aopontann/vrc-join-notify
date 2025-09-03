package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/aopontann/vrc-join-notify/internal/firestore"
	vrc2 "github.com/aopontann/vrc-join-notify/internal/vrc"
	"github.com/bwmarrin/discordgo"
)

type InteractionData struct {
	GuildID string `json:"guild_id"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Options []struct {
		Name    string `json:"name"`
		Options []struct {
			Name  string `json:"name"`
			Type  int    `json:"type"`
			Value string `json:"value"`
		} `json:"options"`
		Type int `json:"type"`
	} `json:"options"`
	Type int `json:"type"`
}

type EventPayloads struct {
	Version       int    `json:"version"`
	ApplicationID string `json:"application_id"`
	Type          int    `json:"type"`
	Event         struct {
		Type      string `json:"type"`
		Timestamp string `json:"timestamp"`
		Data      struct {
			IntegrationType int      `json:"integration_type"`
			Scopes          []string `json:"scopes"`
			User            struct {
				Avatar               string `json:"avatar"`
				AvatarDecorationData any    `json:"avatar_decoration_data"`
				Clan                 any    `json:"clan"`
				Collectibles         any    `json:"collectibles"`
				Discriminator        string `json:"discriminator"`
				DisplayNameStyles    any    `json:"display_name_styles"`
				GlobalName           string `json:"global_name"`
				ID                   string `json:"id"`
				PrimaryGuild         any    `json:"primary_guild"`
				PublicFlags          int    `json:"public_flags"`
				Username             string `json:"username"`
			} `json:"user"`
		} `json:"data"`
	} `json:"event"`
}

func DiscordBotHandler(db *firestore.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 事前認証
		publicKey := os.Getenv("DISCORD_PUBLIC_KEY")
		publicKeyBytes, err := hex.DecodeString(publicKey)
		if err != nil {
			slog.Error("Error decoding hex string: " + err.Error())
			http.Error(w, "Error decoding hex string", http.StatusInternalServerError)
			return
		}
		if !discordgo.VerifyInteraction(r, publicKeyBytes) {
			slog.Error("Invalid request signature")
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

		var eventBody EventPayloads
		if err := json.Unmarshal(body, &eventBody); err != nil {
			http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
			return
		}

		if eventBody.Event.Type == "" {
			fmt.Println("No event type found")
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
		if err != nil {
			http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
			return
		}

		// Webhooks登録時の処理（イベント発生時にリクエストを飛ばすURLを登録する際に必要な処理　初回のみ
		// Interactions Endpoint URL登録時に必要な処理
		if eventBody.Type == 0 || (eventBody.Type == 1 && eventBody.Event.Type == "") {
			pongResp, err := json.Marshal(discordgo.InteractionResponse{
				Type: discordgo.InteractionResponsePong,
			})
			if err != nil {
				http.Error(w, "Error marshalling response", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(pongResp)
			w.WriteHeader(http.StatusOK)
			return
		}

		// アプリをインストールしたときのイベント処理
		if eventBody.Type == 1 && eventBody.Event.Type != "" {
			// DMチャンネルを作成
			ch, err := discord.UserChannelCreate(eventBody.Event.Data.User.ID)
			if err != nil {
				http.Error(w, "Error creating user channel", http.StatusInternalServerError)
				return
			}

			slog.Info("Creating user channel", "channel_id", ch.ID)

			// チャンネルにメッセージを送信
			if _, err := discord.ChannelMessageSend(ch.ID, "Hello! This is a test message from the VRChat Join Notify restapi."); err != nil {
				http.Error(w, "Error sending message", http.StatusInternalServerError)
				return
			}
			slog.Info("Received interaction type 1, sending response")

			// チャンネルIDとユーザIDを保存する処理を追加
			err = db.SaveUserInfo(eventBody.Event.Data.User.ID, ch.ID)
			if err != nil {
				http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
				return
			}
		}

		// スラッシュコマンドの実行時の処理
		if eventBody.Type == 2 {
			var interaction discordgo.Interaction
			if err := json.Unmarshal(body, &interaction); err != nil {
				http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
				return
			}
			var data InteractionData
			// discordgo.Interaction.Data に Name などのフィールドがないため、[]byteに変換して自作の構造体にマッピングする
			jsonData, err := json.Marshal(interaction.Data)
			if err != nil {
				http.Error(w, "Error marshalling request body", http.StatusInternalServerError)
				return
			}
			if err := json.Unmarshal(jsonData, &data); err != nil {
				http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
				return
			}

			// 認証関連処理
			if data.Name == "auth" {
				vrc := vrc2.NewVRC()
				subCmdInfo := data.Options[0]

				// ログイン処理（ユーザ名とパスワードを取得）
				if subCmdInfo.Name == "login" {
					username := subCmdInfo.Options[0].Value
					password := subCmdInfo.Options[1].Value
					token, err := vrc.Login(username, password)
					if err != nil {
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					err = db.SaveUserToken(interaction.User.ID, token)
					if err != nil {
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					// トークンが有効かされたかチェック　メール認証が必要な場合は無効になる
					ok, err := vrc.VerifyAuthToken(token)
					if err != nil {
						ErrorHandler(w, r, err, http.StatusInternalServerError)
					}

					if !ok {
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, "メールに認証コードが送信されました。"); err != nil {
							ErrorHandler(w, r, err, http.StatusInternalServerError)
						}
					} else {
						if err := db.ChangeNotificationed(interaction.User.ID, false); err != nil {
							ErrorHandler(w, r, err, http.StatusInternalServerError)
						}
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, "ログインしました。"); err != nil {
							ErrorHandler(w, r, err, http.StatusInternalServerError)
						}
					}
				}

				if subCmdInfo.Name == "logout" {
					if _, err := discord.ChannelMessageSend(interaction.ChannelID, "未実装"); err != nil {
						http.Error(w, "Error sending message", http.StatusInternalServerError)
						return
					}
				}

				// ログイン処理（2FA）
				if subCmdInfo.Name == "email-code" {
					code := subCmdInfo.Options[0].Value
					// DBからユーザのトークンを取得
					userInfo, err := db.GetUserInfo(interaction.User.ID)
					if err != nil {
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					twoFactorAuthToken, err := vrc.Verify2FA(code, userInfo.Token)
					if err != nil {
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					err = db.SaveUserTwoFactorAuthToken(interaction.User.ID, twoFactorAuthToken)
					if err != nil {
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}

					if _, err := discord.ChannelMessageSend(interaction.ChannelID, "OK"); err != nil {
						http.Error(w, "Error sending message", http.StatusInternalServerError)
						return
					}
				}
			}

			// JOIN通知関連処理
			if data.Name == "join" {
				subCmdInfo := data.Options[0]

				// JOIN通知対象のユーザIDを登録
				if subCmdInfo.Name == "register" {
					url := subCmdInfo.Options[0].Value // https://vrchat.com/home/user/usr_33a8da12-14f4-4225-8711-320471ceb60b
					targetUserID := url[len("https://vrchat.com/home/user/"):]
					err := db.SaveTargetUser(interaction.User.ID, targetUserID)
					if err != nil {
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, err.Error()); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					} else {
						if _, err := discord.ChannelMessageSend(interaction.ChannelID, "OK"); err != nil {
							http.Error(w, "Error sending message", http.StatusInternalServerError)
							return
						}
					}
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
