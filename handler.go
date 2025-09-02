package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

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

func DiscordBotHandler(w http.ResponseWriter, r *http.Request) {
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

	cdb, err := NewDB()
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
		if _, err := discord.ChannelMessageSend(ch.ID, "Hello! This is a test message from the VRChat Join Notify bot."); err != nil {
			http.Error(w, "Error sending message", http.StatusInternalServerError)
			return
		}
		slog.Info("Received interaction type 1, sending response")

		// チャンネルIDとユーザIDを保存する処理を追加
		err = cdb.SaveUserInfo(eventBody.Event.Data.User.ID, ch.ID)
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
			vrc := NewVRC()
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

				err = cdb.SaveUserToken(interaction.User.ID, token)
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
					if err := cdb.ChangeNotificationed(interaction.User.ID, false); err != nil {
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
				userInfo, err := cdb.GetUserInfo(interaction.User.ID)
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

				err = cdb.SaveUserTwoFactorAuthToken(interaction.User.ID, twoFactorAuthToken)
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
				err := cdb.SaveTargetUser(interaction.User.ID, targetUserID)
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

func NotifyHandler(w http.ResponseWriter, r *http.Request) {
	db, err := NewDB()
	if err != nil {
		ErrorHandler(w, r, err, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		ErrorHandler(w, r, err, http.StatusInternalServerError)
		return
	}
	
	vrc := NewVRC()
	
	userInfos, err := db.GetAllUserInfo()
	if err != nil {
		ErrorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	for discordID, userInfo := range userInfos {
		// 初回ログインをしていない場合やターゲットユーザが登録されていない場合はスキップ
		if userInfo.Token == "" || userInfo.TwoFactorAuthToken == "" || userInfo.TargetVRCUserID == "" {
			continue
		}

		// トークンがまだ有効か確認
		ok, err := vrc.VerifyAuthToken(userInfo.Token)
		if err != nil {
			ErrorHandler(w, r, err, http.StatusInternalServerError)
			continue
		}
		if !ok && !userInfo.Notificationed{
			slog.Warn("Auth token is invalid", "discordID", discordID)
			// Discordへの通知
			if _, err := discord.ChannelMessageSend(userInfo.ChannelID, "再ログインしてください。"); err != nil {
				ErrorHandler(w, r, err, http.StatusInternalServerError)
				continue
			}
			// 次回実行時に通知しないようにするための処理
			if err := db.ChangeNotificationed(discordID, true); err != nil {
				ErrorHandler(w, r, err, http.StatusInternalServerError)
			}
			continue
		}

		// ターゲットユーザの情報を取得
		tu, err := vrc.GetUserInfo(userInfo.TargetVRCUserID, userInfo.Token, userInfo.TwoFactorAuthToken)
		if err != nil {
			slog.Error("Failed to get target user info", "error", err)
			continue
		}
		if tu.State == "online" && tu.Status == "join me" && !userInfo.Notificationed {
			// Discordへの通知
			if _, err := discord.ChannelMessageSend(userInfo.ChannelID, tu.DisplayName+" さんがオンラインになりました。"); err != nil {
				ErrorHandler(w, r, err, http.StatusInternalServerError)
				continue
			}

			// 次回実行時に通知しないようにするための処理
			err := db.ChangeNotificationed(discordID, true)
			if err != nil {
				ErrorHandler(w, r, err, http.StatusInternalServerError)
				continue
			}

		}
		// オフラインになった場合、通知フラグをFALSEに戻す
		if tu.State == "online" && userInfo.Notificationed {
			err := db.ChangeNotificationed(discordID, false)
			if err != nil {
				ErrorHandler(w, r, err, http.StatusInternalServerError)
				continue
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error, status int) {
	slog.Error(err.Error())
	http.Error(w, err.Error(), status)
}
