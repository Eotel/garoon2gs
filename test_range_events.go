package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/eotel/garoon2gs/internal/client"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// TestRangeEvents runs test code for entering range events
func TestMain() {
	configDir, err := client.GetConfigDir()
	if err != nil {
		log.Fatal("設定ディレクトリの取得に失敗しました:", err)
	}

	// 明示的に.envファイルを読み込み
	envFile := filepath.Join(configDir, ".env")
	log.Printf("Loading env file from: %s", envFile)
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: .env ファイルの読み込みに失敗しました: %v", err)
	}

	// Google Sheets APIクライアントの初期化
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx,
		option.WithCredentialsFile(filepath.Join(configDir, os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE"))),
		option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		log.Fatal("Google Sheetsクライアントの初期化に失敗しました:", err)
	}

	// テスト用のイベントを作成
	events := []client.Event{
		{
			ID:        "1",
			Subject:   "範囲で休みを入れたテスト",
			EventMenu: "休み",
			Start: client.EventDateTime{
				DateTime: "2025-07-14T00:00:00+09:00",
			},
			End: client.EventDateTime{
				DateTime: "2025-07-18T23:59:00+09:00",
			},
		},
		{
			ID:        "2",
			Subject:   "範囲で出張を入れたテスト",
			EventMenu: "出張",
			Start: client.EventDateTime{
				DateTime: "2025-07-20T00:00:00+09:00",
			},
			End: client.EventDateTime{
				DateTime: "2025-07-22T23:59:00+09:00",
			},
		},
	}

	// テスト用に休暇メニューを読み込み
	holidayMenus, err := TestLoadHolidayMenus()
	if err != nil {
		log.Fatal("休暇メニューの読み込みに失敗しました:", err)
	}

	// ユーザー名は三浦を指定
	userName := "三浦"

	// 予定の書き込み
	if err := SaveToSheet(sheetsService, os.Getenv("SPREADSHEET_ID"), events, holidayMenus, userName); err != nil {
		log.Fatalf("予定書き込みに失敗しました: %v", err)
	}

	log.Printf("テスト予定を正常に書き込みました（%d件）", len(events))
}

// TestLoadHolidayMenus はテスト用に休暇メニューを返します
func TestLoadHolidayMenus() ([]string, error) {
	return []string{"休み"}, nil
}
