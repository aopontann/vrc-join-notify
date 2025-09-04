package vrc

import (
	"net/http"
	"time"
)

type VRC struct {
	Client    http.Client
	BaseURL   string
	UserAgent string
	Cookies   []*http.Cookie
}

type UserInfo struct {
	AgeVerificationStatus string `json:"ageVerificationStatus"`
	AgeVerified           bool   `json:"ageVerified"`
	AllowAvatarCopying    bool   `json:"allowAvatarCopying"`
	Badges                []struct {
		BadgeDescription string `json:"badgeDescription"`
		BadgeID          string `json:"badgeId"`
		BadgeImageURL    string `json:"badgeImageUrl"`
		BadgeName        string `json:"badgeName"`
		Showcased        bool   `json:"showcased"`
	} `json:"badges"`
	Bio                            string        `json:"bio"`
	BioLinks                       []string      `json:"bioLinks"`
	CurrentAvatarImageURL          string        `json:"currentAvatarImageUrl"`
	CurrentAvatarTags              []interface{} `json:"currentAvatarTags"`
	CurrentAvatarThumbnailImageURL string        `json:"currentAvatarThumbnailImageUrl"`
	DateJoined                     string        `json:"date_joined"`
	DeveloperType                  string        `json:"developerType"`
	DisplayName                    string        `json:"displayName"`
	FriendKey                      string        `json:"friendKey"`
	FriendRequestStatus            string        `json:"friendRequestStatus"`
	ID                             string        `json:"id"`
	InstanceID                     string        `json:"instanceId"`
	IsFriend                       bool          `json:"isFriend"`
	LastActivity                   time.Time     `json:"last_activity"`
	LastLogin                      time.Time     `json:"last_login"`
	LastMobile                     interface{}   `json:"last_mobile"`
	LastPlatform                   string        `json:"last_platform"`
	Location                       string        `json:"location"`
	Note                           string        `json:"note"`
	Platform                       string        `json:"platform"`
	ProfilePicOverride             string        `json:"profilePicOverride"`
	ProfilePicOverrideThumbnail    string        `json:"profilePicOverrideThumbnail"`
	Pronouns                       string        `json:"pronouns"`
	State                          string        `json:"state"`
	Status                         string        `json:"status"`
	StatusDescription              string        `json:"statusDescription"`
	Tags                           []string      `json:"tags"`
	TravelingToInstance            string        `json:"travelingToInstance"`
	TravelingToLocation            string        `json:"travelingToLocation"`
	TravelingToWorld               string        `json:"travelingToWorld"`
	UserIcon                       string        `json:"userIcon"`
	WorldID                        string        `json:"worldId"`
}
