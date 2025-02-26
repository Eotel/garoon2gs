package main

import (
	"encoding/csv"
	"fmt"
	"github.com/eotel/garoon2gs/internal/client"
	"log"
	"os"
	"path/filepath"
	"time"
)

// SheetMapping は月とシート名のマッピングを表す構造体です
type SheetMapping struct {
	Month     time.Time
	SheetName string
}

// SheetMapper はイベント日付からシート名を解決するマッパーです
type SheetMapper struct {
	mappings []SheetMapping
}

// NewSheetMapper は新しいSheetMapperインスタンスを作成します
func NewSheetMapper() (*SheetMapper, error) {

	// 環境変数からCSVファイルのパスを取得
	csvPathFromEnv := os.Getenv("SHEET_MAPPING_PATH")
	if csvPathFromEnv == "" {
		return nil, fmt.Errorf("SHEET_MAPPING_PATH environment variable is not set")
	}

	// 実行ファイルのディレクトリを取得
	configDir, err := client.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %v", err)
	}

	// 絶対パスを構築
	csvPath := filepath.Join(configDir, csvPathFromEnv)

	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file %s: %v", csvPath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// ヘッダーを読み飛ばす
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	var mappings []SheetMapping

	// 各行を読み込む
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %v", err)
	}

	for _, record := range records {
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid CSV format: expected 2 columns but got %d", len(record))
		}

		// 月の解析
		month, err := time.Parse("2006-01", record[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse month %q: %v", record[0], err)
		}

		mappings = append(mappings, SheetMapping{
			Month:     month,
			SheetName: record[1],
		})
	}

	log.Printf("Loaded %d sheet mappings from %s", len(mappings), csvPath)
	return &SheetMapper{
		mappings: mappings,
	}, nil
}

// GetSheetName は指定された日付に対応するシート名を返します
// マッピングが存在しない場合はnilを返します
func (sm *SheetMapper) GetSheetName(date time.Time) *string {
	// 年月のみを比較するため、日付部分を初期化
	targetMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())

	for _, m := range sm.mappings {
		if m.Month.Year() == targetMonth.Year() && m.Month.Month() == targetMonth.Month() {
			return &m.SheetName
		}
	}

	log.Printf("Skipping event for %s: no sheet mapping found", date.Format("2006-01"))
	return nil
}

// GetMonthFromSheetName はシート名から対応する年月を返します
// マッピングが存在しない場合はnilを返します
func (sm *SheetMapper) GetMonthFromSheetName(sheetName string) *time.Time {
	for _, m := range sm.mappings {
		if m.SheetName == sheetName {
			return &m.Month
		}
	}
	
	log.Printf("No month mapping found for sheet: %s", sheetName)
	return nil
}
