package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/pkcs12"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config はクライアントの設定を保持する構造体です
type Config struct {
	ConfigDir    string
	BaseURL      string
	Username     string
	Password     string
	CertPath     string
	CertPassword string
}

// GaroonClient はGaroon APIクライアントを表す構造体です
type GaroonClient struct {
	config *Config
	client *http.Client
}

// Event はGaroonの予定を表す構造体です
type Event struct {
	ID        string        `json:"id"`
	Subject   string        `json:"subject"`
	EventMenu string        `json:"eventMenu"`
	Start     EventDateTime `json:"start"`
	End       EventDateTime `json:"end"`
}

// EventDateTime はイベントの開始・終了日時を表す構造体です
type EventDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

func (c *GaroonClient) GetHTTPClient() *http.Client {
	return c.client
}

func (c *GaroonClient) GetBaseURL() string {
	return c.config.BaseURL
}

func (c *GaroonClient) GetUsername() string {
	return c.config.Username
}

func (c *GaroonClient) GetPassword() string {
	return c.config.Password
}

func GetConfigDir() (string, error) {
	// まず実行ファイルのディレクトリを試す
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			return dir, nil
		}
	}

	// 次にカレントディレクトリを試す
	if pwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(pwd, ".env")); err == nil {
			return pwd, nil
		}
	}

	return os.Getwd()
}

// getConfigDir は設定ファイルのディレクトリを取得します
func getConfigDir() (string, error) {
	// まず実行ファイルのディレクトリを試す
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			return dir, nil
		}
	}

	// 次にカレントディレクトリを試す
	if pwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(pwd, ".env")); err == nil {
			return pwd, nil
		}
	}

	// 最後にカレントディレクトリを返す（.envが見つからなくても）
	return os.Getwd()
}

// LoadConfig は環境変数から設定を読み込みます
func LoadConfig() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("設定ディレクトリの取得に失敗しました: %v", err)
	}

	if err := godotenv.Load(filepath.Join(configDir, ".env")); err != nil {
		log.Println("Warning: .env ファイルが見つかりませんでした")
	}

	return &Config{
		ConfigDir:    configDir,
		BaseURL:      os.Getenv("GAROON_BASE_URL"),
		Username:     os.Getenv("GAROON_USERNAME"),
		Password:     os.Getenv("GAROON_PASSWORD"),
		CertPath:     filepath.Join(configDir, os.Getenv("CLIENT_CERT_PATH")),
		CertPassword: os.Getenv("CLIENT_CERT_PASSWORD"),
	}, nil
}

// NewClient は新しいGaroonClientインスタンスを作成します
func NewClient(config *Config) (*GaroonClient, error) {
	var httpClient *http.Client

	if config.CertPath != "" && config.CertPassword != "" {
		pfxData, err := os.ReadFile(config.CertPath)
		if err != nil {
			return nil, fmt.Errorf("証明書の読み込みに失敗しました: %v", err)
		}

		blocks, err := pkcs12.ToPEM(pfxData, config.CertPassword)
		if err != nil {
			return nil, fmt.Errorf("証明書の解析に失敗しました: %v", err)
		}

		var cert tls.Certificate
		for _, b := range blocks {
			if b.Type == "CERTIFICATE" {
				cert.Certificate = append(cert.Certificate, b.Bytes)
			}
			if b.Type == "PRIVATE KEY" {
				cert.PrivateKey, err = x509.ParsePKCS1PrivateKey(b.Bytes)
				if err != nil {
					return nil, fmt.Errorf("秘密鍵の解析に失敗しました: %v", err)
				}
			}
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			},
			Timeout: 10 * time.Second,
		}
	} else {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &GaroonClient{
		config: config,
		client: httpClient,
	}, nil
}

// FetchEvents は指定された期間の予定を取得します
func (c *GaroonClient) FetchEvents(startDate, endDate time.Time, targetUserID string) ([]Event, error) {
	var allEvents []Event
	offset := 0

	// APIリクエストを実行するクロージャ
	fetchPage := func(offset int) ([]Event, bool, error) {
		params := url.Values{}
		params.Add("rangeStart", startDate.Format(time.RFC3339))
		params.Add("rangeEnd", endDate.Format(time.RFC3339))
		params.Add("offset", strconv.Itoa(offset))
		params.Add("limit", "100")
		params.Add("orderBy", "start asc")
		params.Add("target", targetUserID)
		params.Add("targetType", "user")

		reqURL := fmt.Sprintf("%s/api/v1/schedule/events?%s", c.config.BaseURL, params.Encode())
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return nil, false, fmt.Errorf("リクエストの作成に失敗しました: %v", err)
		}

		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.config.Username, c.config.Password)))
		req.Header.Set("X-Cybozu-Authorization", auth)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, false, fmt.Errorf("APIリクエストに失敗しました: %v", err)
		}
		defer resp.Body.Close()

		// ステータスコードチェック
		if resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == 496 { // No Cert
			return nil, false, fmt.Errorf("認証エラー: クライアント証明書が必要です（ステータスコード: %d）", resp.StatusCode)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, false, fmt.Errorf("APIエラー（ステータスコード: %d）: %s", resp.StatusCode, string(body))
		}

		var scheduleResp struct {
			Events  []Event `json:"events"`
			HasNext bool    `json:"hasNext"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&scheduleResp); err != nil {
			return nil, false, fmt.Errorf("JSONのデコードに失敗しました: %v", err)
		}

		return scheduleResp.Events, scheduleResp.HasNext, nil
	}

	// ページング処理
	for {
		events, hasNext, err := fetchPage(offset)
		if err != nil {
			return nil, fmt.Errorf("予定の取得に失敗しました: %v", err)
		}

		allEvents = append(allEvents, events...)

		if !hasNext {
			break
		}
		offset += len(events)
	}

	return allEvents, nil
}
