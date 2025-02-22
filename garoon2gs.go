package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/pkcs12"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Schedule はGaroonのスケジュールを取得するためのクライアントを表す構造体です。
type Schedule struct {
	baseURL      string
	username     string
	password     string
	certPath     string
	certPassword string
	holidayMenus []string
	client       *http.Client
}

// Event はGaroonの予定を表す構造体です。
type Event struct {
	ID        string        `json:"id"`
	Subject   string        `json:"subject"`
	EventMenu string        `json:"eventMenu"`
	Start     EventDateTime `json:"start"`
	End       EventDateTime `json:"end"`
	Location  string        `json:"location,omitempty"`
}

// EventDateTime はイベントの開始・終了日時を表す構造体です。
type EventDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
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

// NewSchedule は新しい Schedule インスタンスを初期化します
func NewSchedule(baseURL, username, password, certPath, certPassword string, holidayMenus []string) (*Schedule, error) {
	var client *http.Client

	if certPath != "" && certPassword != "" {
		// クライアント証明書の設定を行うクロージャ
		setupTLSClient := func() (*http.Client, error) {
			pfxData, err := os.ReadFile(certPath)
			if err != nil {
				return nil, fmt.Errorf("証明書の読み込みに失敗しました: %v", err)
			}

			blocks, err := pkcs12.ToPEM(pfxData, certPassword)
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

			return &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						Certificates: []tls.Certificate{cert},
					},
				},
				Timeout: 10 * time.Second,
			}, nil
		}

		var err error
		client, err = setupTLSClient()
		if err != nil {
			return nil, err
		}
	} else {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &Schedule{
		baseURL:      baseURL,
		username:     username,
		password:     password,
		certPath:     certPath,
		certPassword: certPassword,
		holidayMenus: holidayMenus,
		client:       client,
	}, nil
}

// FetchEvents は指定された期間の予定を取得します
// startDate から endDate までの予定を取得し、ページングがある場合は全て取得します
func (s *Schedule) FetchEvents(startDate, endDate time.Time) ([]Event, error) {
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

		reqURL := fmt.Sprintf("%s/api/v1/schedule/events?%s", s.baseURL, params.Encode())
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return nil, false, err
		}

		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", s.username, s.password)))
		req.Header.Set("X-Cybozu-Authorization", auth)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)

		if resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == 496 { // No Cert
			return nil, false, fmt.Errorf("認証エラー: クライアント証明書が必要な可能性があります")
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Failed to close response body: %v", err)
			}
		}(resp.Body)

		if err != nil {
			return nil, false, err
		}

		var scheduleResp struct {
			Events  []Event `json:"events"`
			HasNext bool    `json:"hasNext"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&scheduleResp); err != nil {
			return nil, false, err
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

// isHoliday は予定が休暇に該当するかを判定します
func (s *Schedule) isHoliday(eventMenu string) bool {
	for _, holiday := range s.holidayMenus {
		if eventMenu == holiday {
			return true
		}
	}
	return false
}

// SaveToSheet は予定をGoogle Sheetsに保存します
func SaveToSheet(serviceAccountFile, spreadsheetID, sheetName string, events []Event, schedule *Schedule) error {
	// スプレッドシートサービスの初期化
	initSheetService := func() (*sheets.Service, error) {
		ctx := context.Background()
		return sheets.NewService(ctx,
			option.WithCredentialsFile(serviceAccountFile),
			option.WithScopes(sheets.SpreadsheetsScope))
	}

	srv, err := initSheetService()
	if err != nil {
		return fmt.Errorf("sheets クライアントの作成に失敗しました: %v", err)
	}

	// シートの存在確認と作成
	ensureSheet := func() error {
		spreadsheet, err := srv.Spreadsheets.Get(spreadsheetID).Do()
		if err != nil {
			return fmt.Errorf("スプレッドシートの取得に失敗しました: %v", err)
		}

		for _, sheet := range spreadsheet.Sheets {
			if sheet.Properties.Title == sheetName {
				return nil
			}
		}

		request := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{Title: sheetName},
				},
			}},
		}
		_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, request).Do()
		return err
	}

	if err := ensureSheet(); err != nil {
		return fmt.Errorf("シートの作成に失敗しました: %v", err)
	}

	// データの書き込み
	writeData := func() error {
		var values [][]interface{}
		header := []interface{}{"ID", "Subject", "EventMenu", "Start", "End", "Location", "IsHoliday"}
		values = append(values, header)

		for _, e := range events {
			row := []interface{}{
				e.ID,
				e.Subject,
				e.EventMenu,
				e.Start.DateTime,
				e.End.DateTime,
				e.Location,
				schedule.isHoliday(e.EventMenu),
			}
			values = append(values, row)
		}

		writeRange := fmt.Sprintf("%s!A1", sheetName)
		valueRange := &sheets.ValueRange{Values: values}
		_, err := srv.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).
			ValueInputOption("USER_ENTERED").Do()
		return err
	}

	if err := writeData(); err != nil {
		return fmt.Errorf("シートの更新に失敗しました: %v", err)
	}

	return nil
}

func main() {
	configDir, err := getConfigDir()
	if err != nil {
		log.Fatal("設定ディレクトリの取得に失敗しました:", err)
	}

	if err := godotenv.Load(filepath.Join(configDir, ".env")); err != nil {
		log.Println("Warning: .env ファイルが見つかりませんでした。")
	}

	// 必須の環境変数を検証する関数
	validateRequiredEnv := func() error {
		required := map[string]string{
			"GAROON_BASE_URL":             os.Getenv("GAROON_BASE_URL"),
			"GAROON_USERNAME":             os.Getenv("GAROON_USERNAME"),
			"GAROON_PASSWORD":             os.Getenv("GAROON_PASSWORD"),
			"SPREADSHEET_ID":              os.Getenv("SPREADSHEET_ID"),
			"GOOGLE_SERVICE_ACCOUNT_FILE": os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE"),
		}

		var missingVars []string
		for key, value := range required {
			if value == "" {
				missingVars = append(missingVars, key)
			}
		}

		if len(missingVars) > 0 {
			return fmt.Errorf("必須の環境変数が設定されていません: %v", missingVars)
		}
		return nil
	}

	if err := validateRequiredEnv(); err != nil {
		log.Fatal(err)
	}

	// 設定の読み込み
	config := struct {
		baseURL        string
		username       string
		password       string
		spreadsheetID  string
		serviceAccount string
		certPath       string
		certPassword   string
	}{
		baseURL:        os.Getenv("GAROON_BASE_URL"),
		username:       os.Getenv("GAROON_USERNAME"),
		password:       os.Getenv("GAROON_PASSWORD"),
		spreadsheetID:  os.Getenv("SPREADSHEET_ID"),
		serviceAccount: filepath.Join(configDir, os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE")),
		certPath:       filepath.Join(configDir, os.Getenv("CLIENT_CERT_PATH")),
		certPassword:   os.Getenv("CLIENT_CERT_PASSWORD"),
	}

	// サービスアカウントファイルの存在確認
	if _, err := os.Stat(config.serviceAccount); err != nil {
		log.Fatalf("サービスアカウントファイルが見つかりません: %v", err)
	}

	// 休暇メニューの取得
	var holidayMenus []string
	if holidayMenusStr := os.Getenv("HOLIDAY_MENUS"); holidayMenusStr != "" {
		if err := json.Unmarshal([]byte(holidayMenusStr), &holidayMenus); err != nil {
			log.Fatal("HOLIDAY_MENUSの解析に失敗しました:", err)
		}
	}

	// Scheduleインスタンスの作成
	schedule, err := NewSchedule(
		config.baseURL,
		config.username,
		config.password,
		config.certPath,
		config.certPassword,
		holidayMenus,
	)
	if err != nil {
		log.Fatalf("Scheduleの初期化に失敗しました: %v", err)
	}

	// 現在から60日後までの予定を取得
	now := time.Now()
	endDate := now.AddDate(0, 0, 60)

	events, err := schedule.FetchEvents(now, endDate)
	if err != nil {
		log.Fatalf("予定の取得に失敗しました: %v", err)
	}

	if len(events) == 0 {
		log.Println("取得した予定が0件でした")
		return
	}

	// スプレッドシートに保存
	err = SaveToSheet(config.serviceAccount, config.spreadsheetID, config.username, events, schedule)
	if err != nil {
		log.Fatalf("スプレッドシートへの保存に失敗しました: %v", err)
	}

	log.Printf("スケジュールデータの書き込みに成功しました（%d件）\n", len(events))
}
