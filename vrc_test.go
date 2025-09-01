package common

import (
	"fmt"
	"testing"
)

func TestGetVRCUserInfo(t *testing.T) {
	vrc := NewVRC()
	auth := ""
	twoFactorAuth := ""
	userID := ""
	userInfo, err := vrc.GetUserInfo(userID, auth, twoFactorAuth)
	if err != nil {
		t.Fatalf("Failed to get user info: %v", err)
	}
	fmt.Println(userInfo)
}

func TestVerifyAuthToken(t *testing.T) {
	auth := ""
	vrc := NewVRC()
	ok, err := vrc.VerifyAuthToken(auth)
	if err != nil {
		t.Fatalf("Failed to verify auth token: %v", err)
	}
	if !ok {
		t.Fatalf("Auth token is not valid")
	}
	t.Log("Auth token is valid")
}

func TestLogin(t *testing.T) {
	username := ""
	passward := ""
	auth, err := NewVRC().Login(username, passward)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	fmt.Println("Auth token:", auth)
}

func TestVerify2FA(t *testing.T) {
	code := ""
	auth := ""
	vrc := NewVRC()
	ok, err := vrc.Verify2FA(code, auth)
	if err != nil {
		t.Fatalf("Failed to verify 2FA: %v", err)
	}
	fmt.Println("2FA verified:", ok)
}
