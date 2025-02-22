package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func main() {
	// 設定ディレクトリの取得
	configDir, err := getConfigDir()
	if err != nil {
		log.Fatal("設定ディレクトリの取得に失敗しました:", err)
	}

	// .envファイルの読み込み
	if err := godotenv.Load(filepath.Join(configDir, ".env")); err != nil {
		log.Fatal(".envファイルの読み込みに失敗しました:", err)
	}

	// 環境変数の取得
	baseURL := os.Getenv("GAROON_BASE_URL")
	username := os.Getenv("GAROON_USERNAME")
	password := os.Getenv("GAROON_PASSWORD")

	if baseURL == "" || username == "" || password == "" {
		log.Fatal("必要な環境変数が設定されていません")
	}

	// クライアントの作成（既存のSchedule構造体から認証関連の処理を流用）
	schedule, err := NewSchedule(
		baseURL,
		username,
		password,
		filepath.Join(configDir, os.Getenv("CLIENT_CERT_PATH")),
		os.Getenv("CLIENT_CERT_PASSWORD"),
		nil,
	)
	if err != nil {
		log.Fatal("クライアントの初期化に失敗しました:", err)
	}

	// ユーザー一覧の取得
	reqURL := fmt.Sprintf("%s/api/v1/base/users", baseURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Fatal("リクエストの作成に失敗しました:", err)
	}

	// Basic認証ヘッダーの設定
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	req.Header.Set("X-Cybozu-Authorization", auth)

	// リクエストの実行
	resp, err := schedule.client.Do(req)
	if err != nil {
		log.Fatal("APIリクエストに失敗しました:", err)
	}
	defer resp.Body.Close()

	// レスポンスのステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("APIエラー（ステータスコード: %d）: %s", resp.StatusCode, string(body))
	}

	// レスポンスの読み取りとJSON解析
	var response struct {
		Users []User `json:"users"`
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		log.Fatal("JSONのデコードに失敗しました:", err)
	}

	// 結果の整形と出力
	prettyJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatal("JSONの整形に失敗しました:", err)
	}

	fmt.Println(string(prettyJSON))
}
