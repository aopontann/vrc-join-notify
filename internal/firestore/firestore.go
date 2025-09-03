package firestore

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

type DB struct {
	Client *firestore.Client
}

func NewDB() (*DB, error) {
	projectID := os.Getenv("PROJECT_ID")
	ctx := context.Background()
	client, err := firestore.NewClientWithDatabase(ctx, projectID, projectID)
	if err != nil {
		log.Fatalln(err)
	}
	return &DB{Client: client}, nil
}

func (db *DB) Close() {
	err := db.Client.Close()
	if err != nil {
		return
	}
}

func (db *DB) SaveUserInfo(discordID string, channelID string) error {
	_, err := db.Client.Collection("users").Doc(discordID).Set(context.Background(), map[string]interface{}{
		"channel_id": channelID,
	})
	return err
}

func (db *DB) SaveUserToken(discordID string, token string) error {
	_, err := db.Client.Collection("users").Doc(discordID).Update(context.Background(), []firestore.Update{
		{
			Path:  "token",
			Value: token,
		},
	})
	return err
}

func (db *DB) SaveUserTwoFactorAuthToken(discordID string, token string) error {
	_, err := db.Client.Collection("users").Doc(discordID).Update(context.Background(), []firestore.Update{
		{
			Path:  "two_factor_auth_token",
			Value: token,
		},
	})
	return err
}

func (db *DB) SaveTargetUser(discordID string, targetID string) error {
	_, err := db.Client.Collection("users").Doc(discordID).Update(context.Background(), []firestore.Update{
		{
			Path:  "target_vrc_user_id",
			Value: targetID,
		},
	})
	return err
}

func (db *DB) GetUserInfo(discordID string) (UserInfo, error) {
	doc, err := db.Client.Collection("users").Doc(discordID).Get(context.Background())
	if err != nil {
		return UserInfo{}, err
	}

	var u UserInfo
	err = doc.DataTo(&u)
	if err != nil {
		return UserInfo{}, err
	}

	return u, nil
}

func (db *DB) GetAllUserInfo() (map[string]UserInfo, error) {
	users := make(map[string]UserInfo)
	iter := db.Client.Collection("users").Documents(context.Background())
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var u UserInfo
		err = doc.DataTo(&u)
		if err != nil {
			return nil, err
		}
		users[doc.Ref.ID] = u
	}
	return users, nil
}

func (db *DB) ChangeNotificationed(discordID string, flag bool) error {
	_, err := db.Client.Collection("users").Doc(discordID).Update(context.Background(), []firestore.Update{
		{
			Path:  "notificationed",
			Value: flag,
		},
	})
	return err
}
