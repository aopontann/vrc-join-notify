package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/aopontann/vrc-join-notify/internal/firestore"
	"github.com/aopontann/vrc-join-notify/internal/handler"
	godotenv "github.com/joho/godotenv"
)

// DiscordWebhook をローカルで動作確認するためのエンドポイント
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

	db, err := firestore.NewDB()
	if err != nil {
		slog.Error("failed to connect to firestore: " + err.Error())
		return
	}

	http.HandleFunc("/bot", handler.DiscordBotHandler(db))
	http.HandleFunc("/notify", handler.NotifyHandler(db))

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
