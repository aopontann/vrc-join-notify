package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	ListCommand()
	// BulkCommand()
}

func ListCommand() {
	godotenv.Load(".env.dev")
	appID := os.Getenv("DISCORD_APP_ID")

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	cmds, err := discord.ApplicationCommands(appID, "")
	if err != nil {
		panic(err)
	}
	for _, cmd := range cmds {
		fmt.Println(cmd.ID, cmd.Name)
	}
}

func BulkCommand() {
	godotenv.Load(".env.dev")
	appID := os.Getenv("DISCORD_APP_ID")

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	_, err = discord.ApplicationCommandCreate(appID, "", &discordgo.ApplicationCommand{
		Name:        "auth",
		Description: "認証処理",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "login",
				Description: "ログイン",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "username",
						Description: "ユーザ名",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
					{
						Name:        "password",
						Description: "パスワード",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "email-code",
				Description: "emailotpによる2FA認証",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "code",
						Description: "コード",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "logout",
				Description: "ログアウト",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options:     []*discordgo.ApplicationCommandOption{},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	_, err = discord.ApplicationCommandCreate(appID, "", &discordgo.ApplicationCommand{
		Name:        "join",
		Description: "JOIN通知対象のユーザ処理",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "register",
				Description: "JOIN通知対象のユーザを登録",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "url",
						Description: "ユーザー情報のURL",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	// _, err = discord.ApplicationCommandBulkOverwrite(appID, "", []*discordgo.ApplicationCommand{
	// 	{
	// 		Name:        "auth",
	// 		Description: "認証処理",
	// 		Options: []*discordgo.ApplicationCommandOption{
	// 			{
	// 				Name:        "login",
	// 				Description: "ログイン",
	// 				Type:        discordgo.ApplicationCommandOptionSubCommand,
	// 				Options: []*discordgo.ApplicationCommandOption{
	// 					{
	// 						Name:        "username",
	// 						Description: "ユーザ名",
	// 						Type:        discordgo.ApplicationCommandOptionString,
	// 						Required:    true,
	// 					},
	// 					{
	// 						Name:        "password",
	// 						Description: "パスワード",
	// 						Type:        discordgo.ApplicationCommandOptionString,
	// 						Required:    true,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name:        "2FA",
	// 				Description: "emailotpによる2FA認証",
	// 				Type:        discordgo.ApplicationCommandOptionSubCommand,
	// 				Options: []*discordgo.ApplicationCommandOption{
	// 					{
	// 						Name:        "code",
	// 						Description: "コード",
	// 						Type:        discordgo.ApplicationCommandOptionString,
	// 						Required:    true,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name:        "logout",
	// 				Description: "ログアウト",
	// 				Type:        discordgo.ApplicationCommandOptionSubCommand,
	// 				Options:     []*discordgo.ApplicationCommandOption{},
	// 			},
	// 		},
	// 	},
	// 	{
	// 		Name:        "join",
	// 		Description: "JOIN通知対象のユーザ処理",
	// 		Options: []*discordgo.ApplicationCommandOption{
	// 			{
	// 				Name:        "register",
	// 				Description: "JOIN通知対象のユーザを登録",
	// 				Type:        discordgo.ApplicationCommandOptionSubCommand,
	// 				Options: []*discordgo.ApplicationCommandOption{
	// 					{
	// 						Name:        "url",
	// 						Description: "ユーザー情報のURL",
	// 						Type:        discordgo.ApplicationCommandOptionString,
	// 						Required:    true,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	panic(err)
	// }
}
