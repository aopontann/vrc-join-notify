package discord

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env.dev"); err != nil {
		panic(err)
	}
	m.Run()
}

func TestChannelMessageSend(t *testing.T) {
	discord, err := New()
	if err != nil {
		t.Fatal(err)
	}

	c := `
	テスト
	`

	_, err = discord.ChannelMessageSend("1407728142705233991", c)
	if err != nil {
		t.Fatal(err)
	}
}
