package handler

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/aopontann/vrc-join-notify/internal/firestore"
	vrc2 "github.com/aopontann/vrc-join-notify/internal/vrc"
	"github.com/bwmarrin/discordgo"
)

func NotifyHandler(db *firestore.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
		if err != nil {
			ErrorHandler(w, r, err, http.StatusInternalServerError)
			return
		}

		vrc := vrc2.NewVRC()

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

			// トークンが無効な場合はDiscordに通知
			if !ok {
				// ログは毎回表示
				slog.Warn("Auth token is invalid", "discordID", discordID)

				// ただし、Discordへの通知は一度だけにする
				// 通知フラグがFALSEの場合のみ通知を行い、通知後にTRUEに変更する
				if !userInfo.Notificationed {
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
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error, status int) {
	slog.Error(err.Error())
	http.Error(w, err.Error(), status)
}
