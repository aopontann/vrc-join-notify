package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	c "github.com/aopontann/vrc-join-notify"
)

func main() {
	// Cloud Logging用のログ設定
	ops := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				a.Key = "severity"
				level := a.Value.Any().(slog.Level)
				if level == slog.LevelWarn {
					a.Value = slog.StringValue("WARNING")
				}
			}

			return a
		},
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &ops))
	slog.SetDefault(logger)

	if os.Getenv("ENV") != "prod" {
		slog.Debug("Loading environmental variables...")
		if err := godotenv.Load(".env.dev"); err != nil {
			slog.Error("failed to load env variables: " + err.Error())
			return
		}
	}

	http.HandleFunc("/", Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Debug(fmt.Sprintf("Listening on port %s", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("something went terribly wrong: " + err.Error())
		return
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	db, err := c.NewDB()
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

	userInfos, err := db.GetAllUserInfo()
	if err != nil {
		ErrorHandler(w, r, err, http.StatusInternalServerError)
		return
	}

	vrc := c.NewVRC()
	for discordID, userInfo := range userInfos {
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
