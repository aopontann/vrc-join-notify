package vrc

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func NewVRC() *VRC {
	return &VRC{
		Client:    http.Client{},
		BaseURL:   "https://api.vrchat.cloud/api/1",
		UserAgent: os.Getenv("USER_AGENT"),
		Cookies:   nil,
	}
}

// VerifyAuthToken は現在提供されている認証トークンが有効かどうかを確認する。
func (v *VRC) VerifyAuthToken(token string) (bool, error) {
	path := "/auth"
	req, err := http.NewRequest("GET", v.BaseURL+path, nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return false, err
	}
	req.Header.Add("user-agent", v.UserAgent)

	// 認証情報をセット
	req.Header.Add("Cookie", "auth="+token)

	// リクエスト実行
	resp, err := v.Client.Do(req)
	if err != nil {
		slog.Error("Failed to execute request", "error", err)
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func (v *VRC) Login(username, password string) (string, error) {
	path := "/auth/user"
	req, err := http.NewRequest("GET", v.BaseURL+path, nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)
		return "", err
	}

	// 認証情報をセット
	req.SetBasicAuth(username, password)
	req.Header.Add("user-agent", v.UserAgent)
	// リクエスト実行
	resp, err := v.Client.Do(req)
	if err != nil {
		slog.Error("Failed to execute request", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		fmt.Println(cookie)
		slog.Info("Received cookie", "name", cookie.Name, "value", cookie.Value)
		if cookie.Name == "auth" {
			slog.Info("Found auth cookie", "value", cookie.Value)
			return cookie.Value, nil
		}
	}

	slog.Info("No auth cookie found in response")
	return "", nil
}

func (v *VRC) Verify2FA(code string, auth string) (string, error) {
	path := "/auth/twofactorauth/emailotp/verify"
	bodyStr := `{"code":"` + code + `"}`
	body := strings.NewReader(bodyStr)
	req, _ := http.NewRequest("POST", v.BaseURL+path, body)
	req.Header.Add("user-agent", v.UserAgent)
	req.Header.Add("Cookie", "auth="+auth)
	req.Header.Add("Content-Type", "application/json")

	resp, err := v.Client.Do(req)
	if err != nil {
		slog.Error("Failed to execute request", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to verify 2FA", "status", resp.StatusCode)
		return "", fmt.Errorf("failed to verify 2FA, status code: %d", resp.StatusCode)
	}

	for _, cookie := range resp.Cookies() {
		slog.Info("Received cookie", "name", cookie.Name, "value", cookie.Value)
		if cookie.Name == "twoFactorAuth" {
			slog.Info("Found auth cookie", "value", cookie.Value)
			return cookie.Value, nil
		}
	}

	slog.Info("No twoFactorAuth cookie found in response")
	return "", nil
}

func (v *VRC) GetUserInfo(userID string, auth string, twoFactorAuth string) (UserInfo, error) {
	path := "/users/" + userID
	req, _ := http.NewRequest("GET", v.BaseURL+path, nil)
	req.Header.Add("user-agent", v.UserAgent)
	req.Header.Add("Cookie", "auth="+auth+";twoFactorAuth="+twoFactorAuth)

	resp, err := v.Client.Do(req)
	if err != nil {
		slog.Error("Failed to execute request", "error", err)
		return UserInfo{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return UserInfo{}, err
	}

	fmt.Println(string(body))

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		slog.Error("Failed to unmarshal user info", "error", err)
		return UserInfo{}, err
	}

	return userInfo, nil
}
