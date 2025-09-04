package discord

import (
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Session   *discordgo.Session
	publicKey string
}

func New() (*Discord, error) {
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		return nil, err
	}
	return &Discord{
		Session:   discord,
		publicKey: os.Getenv("DISCORD_PUBLIC_KEY"),
	}, nil
}

func (d *Discord) VerifyInteraction(r *http.Request) (bool, error) {
	publicKeyBytes, err := hex.DecodeString(d.publicKey)
	if err != nil {
		slog.Error("Error decoding hex string: " + err.Error())
		return false, err
	}

	return discordgo.VerifyInteraction(r, publicKeyBytes), nil
}

func (d *Discord) GetEventInfo(b []byte) (int, string, string, *InteractionData, error) {
	var eventBody EventPayloads
	if err := json.Unmarshal(b, &eventBody); err != nil {
		return InternalError, "", "", nil, err
	}

	if eventBody.Type == 0 {
		return WebhooksResist, "", "", nil, nil
	}

	if eventBody.Type == 1 && eventBody.Event.Type == "" {
		return InteractionsEndpointResist, "", "", nil, nil
	}

	if eventBody.Type == 1 && eventBody.Event.Type != "" {
		// アプリをインストールしたユーザのID
		uid := eventBody.Event.Data.User.ID
		return ApplicationInstall, uid, "", nil, nil
	}

	if eventBody.Type == 2 {
		uid := eventBody.Event.Data.User.ID

		var interaction discordgo.Interaction
		if err := json.Unmarshal(b, &interaction); err != nil {
			return InternalError, "", "", nil, err
		}
		var data InteractionData
		// discordgo.Interaction.Data に Name などのフィールドがないため、[]byteに変換して自作の構造体にマッピングする
		jsonData, err := json.Marshal(interaction.Data)
		if err != nil {
			return InternalError, "", "", nil, err
		}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return InternalError, "", "", nil, err
		}

		cid := interaction.ChannelID
		return SlashCommand, uid, cid, &data, nil
	}

	// このリターンにたどり着くことはないが、エラーが表示されるため実装
	return InternalError, "", "", nil, nil
}

func (d *Discord) UserChannelCreate(uid string) (*discordgo.Channel, error) {
	ch, err := d.Session.UserChannelCreate(uid)
	return ch, err
}

func (d *Discord) ChannelMessageSend(cid string, content string) (*discordgo.Message, error) {
	m, err := d.Session.ChannelMessageSend(cid, content)
	return m, err
}
