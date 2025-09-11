package firestore

import (
	"fmt"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env.dev"); err != nil {
		panic(err)
	}
	m.Run()
}

func TestNewDB(t *testing.T) {
	client, err := NewDB()
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}
	if client == nil {
		t.Fatal("Firestore client is nil")
	}
	defer client.Close()
}

func TestSaveUserInfo(t *testing.T) {
	client, err := NewDB()
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}

	discordID := "test_discord_id"
	channelID := "test_channel_id"
	err = client.SaveUserInfo(discordID, channelID)
	if err != nil {
		t.Fatalf("Failed to save user info: %v", err)
	}
	defer client.Close()
}

func TestSaveUserToken(t *testing.T) {
	client, err := NewDB()
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}

	discordID := "test_discord_id"
	token := "test_token"
	err = client.SaveUserToken(discordID, token)
	if err != nil {
		t.Fatalf("Failed to save user info: %v", err)
	}
	defer client.Close()
}

func TestSaveTargetUser(t *testing.T) {
	client, err := NewDB()
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}

	discordID := "test_discord_id"
	targetID := "target_vrc_user_id"
	err = client.SaveTargetUser(discordID, targetID)
	if err != nil {
		t.Fatalf("Failed to save user info: %v", err)
	}
	defer client.Close()
}

func TestGetUserInfo(t *testing.T) {
	client, err := NewDB()
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}

	discordID := "test_discord_id"
	userInfo, err := client.GetUserInfo(discordID)
	if err != nil {
		t.Fatalf("Failed to save user info: %v", err)
	}
	fmt.Println(userInfo)
	defer client.Close()
}

func TestGetAllUserInfo(t *testing.T) {
	client, err := NewDB()
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}

	users, err := client.GetAllUserInfo()
	if err != nil {
		t.Fatalf("Failed to get all user info: %v", err)
	}

	for discordID, userInfo := range users {
		fmt.Printf("Discord ID: %s, User Info: %+v\n", discordID, userInfo)
	}
	defer client.Close()
}
