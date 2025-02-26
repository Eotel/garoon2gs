package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/eotel/garoon2gs/internal/client"
	"github.com/eotel/garoon2gs/internal/mapping"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// コマンドラインオプションの設定
	showVersion := flag.Bool("version", false, "バージョン情報を表示")
	flag.Parse()

	// バージョン情報の表示
	if *showVersion {
		fmt.Printf("Garoon2GS version %s, commit %s, built at %s\n", version, commit, date)
		return
	}

	configDir, err := client.GetConfigDir()
	if err != nil {
		log.Fatal("設定ディレクトリの取得に失敗しました:", err)
	}

	if err := godotenv.Load(filepath.Join(configDir, ".env")); err != nil {
		log.Println("Warning: .env ファイルが見つかりませんでした。")
	}

	// 必須の環境変数を検証
	if err := validateRequiredEnv(); err != nil {
		log.Fatal(err)
	}

	// クライアントの設定を読み込み
	config, err := client.LoadConfig()
	if err != nil {
		log.Fatal("設定の読み込みに失敗しました:", err)
	}

	// Garoonクライアントの初期化
	garoonClient, err := client.NewClient(config)
	if err != nil {
		log.Fatal("Garoonクライアントの初期化に失敗しました:", err)
	}

	// ユーザーマッピングの読み込み
	userMappings, err := mapping.LoadUserMapping(configDir)
	if err != nil {
		log.Fatal("ユーザーマッピングの読み込みに失敗しました:", err)
	}

	// Google Sheets APIクライアントの初期化
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithCredentialsFile(filepath.Join(configDir, os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE"))),
		option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		log.Fatal("Google Sheetsクライアントの初期化に失敗しました:", err)
	}

	// 休暇メニューの読み込み
	holidayMenus, err := loadHolidayMenus()
	if err != nil {
		log.Fatal("休暇メニューの読み込みに失敗しました:", err)
	}

	// 期間の設定
	startDate, endDate := calculateDateRange()
	log.Printf("取得期間: %s から %s まで", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 各ユーザーの予定を取得して保存
	for _, userMapping := range userMappings {
		log.Printf("ユーザーID %s の予定を取得します", userMapping.UserID)

		events, err := garoonClient.FetchEvents(startDate, endDate, userMapping.UserID)
		if err != nil {
			log.Printf("警告: ユーザーID %s の予定取得に失敗しました: %v", userMapping.UserID, err)
			continue
		}

		if len(events) == 0 {
			log.Printf("ユーザーID %s の予定は0件でした", userMapping.UserID)
			continue
		}

		// 予定の書き込み
		if err := SaveToSheet(sheetsService, os.Getenv("SPREADSHEET_ID"), events, holidayMenus, userMapping.HeaderName); err != nil {
			log.Printf("警告: ユーザーID %s の予定書き込みに失敗しました: %v", userMapping.UserID, err)
			continue
		}

		log.Printf("ユーザーID %s の予定を正常に書き込みました（%d件）", userMapping.UserID, len(events))
	}
}

// validateRequiredEnv は必須の環境変数を検証します
func validateRequiredEnv() error {
	required := map[string]string{
		"GAROON_BASE_URL":             os.Getenv("GAROON_BASE_URL"),
		"GAROON_USERNAME":             os.Getenv("GAROON_USERNAME"),
		"GAROON_PASSWORD":             os.Getenv("GAROON_PASSWORD"),
		"SPREADSHEET_ID":              os.Getenv("SPREADSHEET_ID"),
		"GOOGLE_SERVICE_ACCOUNT_FILE": os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE"),
		"USER_MAPPING_PATH":           os.Getenv("USER_MAPPING_PATH"),
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

// loadHolidayMenus は休暇メニューを環境変数から読み込みます
func loadHolidayMenus() ([]string, error) {
	var holidayMenus []string
	if holidayMenusStr := os.Getenv("HOLIDAY_MENUS"); holidayMenusStr != "" {
		if err := json.Unmarshal([]byte(holidayMenusStr), &holidayMenus); err != nil {
			return nil, fmt.Errorf("HOLIDAY_MENUSの解析に失敗しました: %v", err)
		}
	}
	return holidayMenus, nil
}

// calculateDateRange は取得対象の期間を計算します
func calculateDateRange() (time.Time, time.Time) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	endYear, endMonth := now.Year(), now.Month()+3
	if endMonth > 12 {
		endYear++
		endMonth = endMonth - 12
	}
	endDate := time.Date(endYear, endMonth+1, 1, 0, 0, 0, 0, time.Local).Add(-time.Second)

	return startDate, endDate
}

// SaveToSheet は予定をスプレッドシートに保存します
func SaveToSheet(srv *sheets.Service, spreadsheetID string, events []client.Event, holidayMenus []string, userName string) error {
	// スケジュール書き込み用のインスタンスを作成
	writer, err := NewScheduleWriter()
	if err != nil {
		return fmt.Errorf("schedule writerの作成に失敗しました: %v", err)
	}
	writer.holidayMenus = holidayMenus
	writer.name = userName // ユーザー名を設定

	// イベントを日付でグループ化
	eventsByDate := make(map[string]map[int][]client.Event)
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
			continue
		}

		if eventsByDate[*targetSheet] == nil {
			eventsByDate[*targetSheet] = make(map[int][]client.Event)
		}

		day := eventTime.Day()
		eventsByDate[*targetSheet][day] = append(eventsByDate[*targetSheet][day], e)
	}

	// シートごとに書き込み
	for sheetName, dailyEvents := range eventsByDate {
		err = writer.WriteSchedule(srv, spreadsheetID, sheetName, dailyEvents)
		if err != nil {
			return fmt.Errorf("シート %s の更新に失敗しました: %v", sheetName, err)
		}
	}

	return nil
}
