package firestore

type UserInfo struct {
	ChannelID          string `firestore:"channel_id,omitempty"`
	TargetVRCUserID    string `firestore:"target_vrc_user_id,omitempty"`
	Token              string `firestore:"token,omitempty"`
	TwoFactorAuthToken string `firestore:"two_factor_auth_token,omitempty"`
	Notificationed     bool   `firestore:"notificationed,omitempty"`
}
