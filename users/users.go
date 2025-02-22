package users

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// User represents a Garoon user
type User struct {
	ID                  string `json:"id"`
	Code                string `json:"code"`
	Name                string `json:"name"`
	Email               string `json:"email,omitempty"`
	Status              string `json:"status"`
	PrimaryOrganization struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"primaryOrganization"`
}

// ListUsers はユーザー一覧を取得する関数です
func ListUsers(client *http.Client, baseURL, username, password string) ([]User, error) {
	reqURL := fmt.Sprintf("%s/api/v1/base/users", baseURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %v", err)
	}

	// Basic認証ヘッダーの設定
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	req.Header.Set("X-Cybozu-Authorization", auth)

	// リクエストの実行
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストに失敗しました: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスのステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("APIエラー（ステータスコード: %d）: %s", resp.StatusCode, string(body))
	}

	// レスポンスの読み取りとJSON解析
	var response struct {
		Users []User `json:"users"`
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("JSONのデコードに失敗しました: %v", err)
	}

	return response.Users, nil
}

// PrintUsers はユーザー一覧を整形して出力する関数です
func PrintUsers(users []User) error {
	// 結果の整形と出力
	prettyJSON, err := json.MarshalIndent(struct {
		Users []User `json:"users"`
	}{users}, "", "  ")
	if err != nil {
		return fmt.Errorf("JSONの整形に失敗しました: %v", err)
	}

	fmt.Println(string(prettyJSON))
	return nil
}
