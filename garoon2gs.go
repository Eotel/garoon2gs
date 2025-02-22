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
	"strings"
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
// FetchEvents を修正
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
			return nil, false, fmt.Errorf("リクエストの作成に失敗しました: %v", err)
		}

		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", s.username, s.password)))
		req.Header.Set("X-Cybozu-Authorization", auth)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, false, fmt.Errorf("APIリクエストに失敗しました: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("レスポンスのクローズに失敗しました: %v", err)
			}
		}(resp.Body)

		// ステータスコードチェックを先に行う
		if resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == 496 { // No Cert
			return nil, false, fmt.Errorf("認証エラー: クライアント証明書が必要です（ステータスコード: %d）", resp.StatusCode)
		}

		if resp.StatusCode != http.StatusOK {
			// レスポンスボディの最初の部分だけを取得（HTMLの場合は表示しない）
			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)
			if strings.HasPrefix(bodyStr, "<!DOCTYPE") || strings.HasPrefix(bodyStr, "<html") {
				return nil, false, fmt.Errorf("APIエラー（ステータスコード: %d）: クライアント証明書が必要です", resp.StatusCode)
			}
			// JSONの場合はそのまま表示（最大100文字まで）
			if len(bodyStr) > 100 {
				bodyStr = bodyStr[:100] + "..."
			}
			return nil, false, fmt.Errorf("APIエラー（ステータスコード: %d）: %s", resp.StatusCode, bodyStr)
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
func SaveToSheet(serviceAccountFile, spreadsheetID string, events []Event, schedule *Schedule) error {
	// スプレッドシートサービスの初期化
	ctx := context.Background()
	srv, err := sheets.NewService(ctx,
		option.WithCredentialsFile(serviceAccountFile),
		option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return fmt.Errorf("sheets クライアントの作成に失敗しました: %v", err)
	}

	// スプレッドシートの情報を取得
	spreadsheet, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("スプレッドシートの取得に失敗しました: %v", err)
	}

	// 既存のシート名を取得
	existingSheets := make(map[string]bool)
	for _, sheet := range spreadsheet.Sheets {
		existingSheets[sheet.Properties.Title] = true
	}

	// イベントを日付でグループ化
	eventsByDate := make(map[string]map[int][]Event)
	for _, e := range events {
		eventTime, err := time.Parse(time.RFC3339, e.Start.DateTime)
		if err != nil {
			log.Printf("イベントの日時解析に失敗しました: %v", err)
			continue
		}

		// シート名を取得
		sheetMapper, err := NewSheetMapper()
		if err != nil {
			return fmt.Errorf("sheet mapperの作成に失敗しました: %v", err)
		}

		targetSheet := sheetMapper.GetSheetName(eventTime)
		if targetSheet == nil {
			// マッピングにない月はスキップ
			continue
		}

		// シートが存在しない場合はスキップ
		if !existingSheets[*targetSheet] {
			log.Printf("シート %s が存在しないためスキップします", *targetSheet)
			continue
		}

		// シートごとのマップを初期化
		if eventsByDate[*targetSheet] == nil {
			eventsByDate[*targetSheet] = make(map[int][]Event)
		}

		// 日付でグループ化
		day := eventTime.Day()
		eventsByDate[*targetSheet][day] = append(eventsByDate[*targetSheet][day], e)
	}

	// スケジュール書き込み用のインスタンスを作成
	writer, err := NewScheduleWriter()
	if err != nil {
		return fmt.Errorf("schedule writerの作成に失敗しました: %v", err)
	}
	writer.holidayMenus = schedule.holidayMenus

	// シートごとに書き込み
	for sheetName, dailyEvents := range eventsByDate {
		// データの書き込み
		err = writer.WriteSchedule(srv, spreadsheetID, sheetName, dailyEvents)
		if err != nil {
			return fmt.Errorf("シート %s の更新に失敗しました: %v", sheetName, err)
		}
	}

	return nil
}

// ensureSheet はシートの存在確認と作成を行います
func ensureSheet(srv *sheets.Service, spreadsheetID, sheetName string) error {
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

	// 現在の年月を取得
	now := time.Now()
	currentYear := now.Year()
	currentMonth := now.Month()

	// 現在の月の初日
	startDate := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.Local)

	// 2ヶ月後の月末日を計算
	endYear := currentYear
	endMonth := currentMonth + 3
	if endMonth > 12 {
		endYear++
		endMonth = endMonth - 12
	}
	endDate := time.Date(endYear, endMonth+1, 1, 0, 0, 0, 0, time.Local).Add(-time.Second)

	log.Printf("取得期間: %s から %s まで", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	events, err := schedule.FetchEvents(startDate, endDate)
	if err != nil {
		log.Fatalf("予定の取得に失敗しました: %v", err)
	}

	if len(events) == 0 {
		log.Println("取得した予定が0件でした")
		return
	}

	// スプレッドシートに保存
	err = SaveToSheet(config.serviceAccount, config.spreadsheetID, events, schedule)
	if err != nil {
		log.Fatalf("スプレッドシートへの保存に失敗しました: %v", err)
	}

	log.Printf("スケジュールデータの書き込みに成功しました（%d件）\n", len(events))
}
