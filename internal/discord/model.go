package discord

const (
	WebhooksResist             = 0 // Webhooks登録時のイベント
	InteractionsEndpointResist = 1 // Interactions Endpoint URL登録時のイベント
	ApplicationInstall         = 2 // アプリをインストールしたときのイベント
	SlashCommand               = 3 // スラッシュコマンドの実行時のイベント
	InternalError              = 4
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
