package main

import (
	"encoding/json"
	"fmt"
	"github.com/eotel/garoon2gs/internal/client"
	"google.golang.org/api/sheets/v4"
	"log"
	"os"
	"strconv"
	"time"
)

// ScheduleWriter はスケジュールをスプレッドシートに書き込むための構造体です
type ScheduleWriter struct {
	headerRow    int
	dateCol      string
	name         string
	holidayMenus []string
	outingMenus  []string // 外出、出張などの特殊な出勤
	normalPlace  string   // 通常の勤務地（"渋谷"）
	nameCol      string
}

// NewScheduleWriter は新しい ScheduleWriter インスタンスを作成します
func NewScheduleWriter() (*ScheduleWriter, error) {
	// 既存の環境変数チェック
	headerRowStr := os.Getenv("HEADER_ROW")
	if headerRowStr == "" {
		return nil, fmt.Errorf("HEADER_ROW environment variable is not set")
	}

	headerRow, err := strconv.Atoi(headerRowStr)
	if err != nil {
		return nil, fmt.Errorf("invalid HEADER_ROW value: %v", err)
	}

	dateCol := os.Getenv("DATE_COL")
	if dateCol == "" {
		return nil, fmt.Errorf("DATE_COL environment variable is not set")
	}

	// name := os.Getenv("NAME")
	// if name == "" {
	// 	return nil, fmt.Errorf("NAME environment variable is not set")
	// }

	// 新しい環境変数の読み込み
	var holidayMenus []string
	if holidayMenusStr := os.Getenv("HOLIDAY_MENUS"); holidayMenusStr != "" {
		if err := json.Unmarshal([]byte(holidayMenusStr), &holidayMenus); err != nil {
			return nil, fmt.Errorf("failed to parse HOLIDAY_MENUS: %v", err)
		}
	}

	var outingMenus []string
	if outingMenusStr := os.Getenv("OUTING_MENUS"); outingMenusStr != "" {
		if err := json.Unmarshal([]byte(outingMenusStr), &outingMenus); err != nil {
			return nil, fmt.Errorf("failed to parse OUTING_MENUS: %v", err)
		}
	}

	normalPlace := os.Getenv("NORMAL_PLACE")
	if normalPlace == "" {
		normalPlace = "渋谷" // デフォルト値
	}

	return &ScheduleWriter{
		headerRow:    headerRow,
		dateCol:      dateCol,
		name:         "", // SaveToSheet()で設定されるため空文字で初期化
		holidayMenus: holidayMenus,
		outingMenus:  outingMenus,
		normalPlace:  normalPlace,
	}, nil
}

// findNameColumn はヘッダー行から名前の列を特定します
func (w *ScheduleWriter) findNameColumn(headerValues []interface{}) (string, error) {
	log.Printf("Searching for name '%s' in header values: %v", w.name, headerValues)
	for i, value := range headerValues {
		if str, ok := value.(string); ok {
			log.Printf("Checking header column %d: '%s'", i, str)
			if str == w.name {
				col := columnIndexToName(i)
				log.Printf("Found name in column %s", col)
				return col, nil
			}
		}
	}
	return "", fmt.Errorf("column for name %q not found in header row", w.name)
}

// getLastDateRow はDATE列の最後の日付を探して、最終行を特定します
func (w *ScheduleWriter) getLastDateRow(srv *sheets.Service, spreadsheetID, sheetName string) (int, error) {
	dateRange := fmt.Sprintf("%s!%s%d:%s%d", sheetName, w.dateCol, w.headerRow+1, w.dateCol, 100)
	log.Printf("Reading date column range: %s", dateRange)

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, dateRange).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to read date column: %v", err)
	}

	log.Printf("Date column values: %+v", resp.Values)

	if len(resp.Values) == 0 {
		return 0, fmt.Errorf("no data found in date column")
	}

	// 数値が入っている最後の行を探す
	lastRow := w.headerRow
	for i, row := range resp.Values {
		if len(row) == 0 {
			continue
		}

		// 数値として解釈できる値のみを考慮
		switch v := row[0].(type) {
		case float64:
			if v > 0 {
				lastRow = w.headerRow + i + 1
			}
		case int:
			if v > 0 {
				lastRow = w.headerRow + i + 1
			}
		case string:
			if num, err := strconv.Atoi(v); err == nil && num > 0 {
				lastRow = w.headerRow + i + 1
			}
		}
	}

	log.Printf("Found last date row at: %d", lastRow)
	return lastRow, nil
}

// getCellPosition は指定された日付に対応するセルの位置を返します
func (w *ScheduleWriter) getCellPosition(date time.Time) (row int, col string, err error) {
	if w.nameCol == "" {
		return 0, "", fmt.Errorf("name column is not initialized")
	}

	// 日付から行番号を計算
	day := date.Day()
	row = w.headerRow + day

	return row, w.nameCol, nil
}

// determineEventStatus はイベントの状態を判定します
func (w *ScheduleWriter) determineEventStatus(events []client.Event) string {
	if len(events) == 0 {
		// 予定がない場合は通常の勤務地を返す
		return "渋谷"
	}

	// 1. 休み判定が一つでもあるかチェック
	for _, event := range events {
		for _, holiday := range w.holidayMenus {
			if event.EventMenu == holiday {
				return "週休" // 休み判定があれば必ず"週休"を返す
			}
		}
	}

	// 2. OUTING_MENUSに該当するものがあるかチェック
	for _, event := range events {
		for _, outingMenu := range w.outingMenus {
			if event.EventMenu == outingMenu {
				return "外出"
			}
		}
	}

	// 3. それ以外の場合は "渋谷" を返す
	return "渋谷"
}

// columnIndexToName は0-based indexをA1記法の列名に変換します
func columnIndexToName(index int) string {
	name := ""
	index++ // 1-basedに変換（Excel/Googleスプレッドシートの列は1から始まるため）

	for index > 0 {
		index-- // 0-basedに戻す（'A'から始まるため）
		name = string('A'+byte(index%26)) + name
		index = index / 26
	}
	return name
}

// WriteSchedule は指定されたシートにスケジュールを書き込みます
func (w *ScheduleWriter) WriteSchedule(srv *sheets.Service, spreadsheetID, sheetName string, monthlyEvents map[int][]client.Event) error {
	// まずヘッダー行から名前の列を特定
	headerRange := fmt.Sprintf("%s!%d:%d", sheetName, w.headerRow, w.headerRow)
	log.Printf("Reading header row from range: %s", headerRange)

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, headerRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read header row: %v", err)
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return fmt.Errorf("header row is empty")
	}

	// 名前の列を特定
	w.nameCol, err = w.findNameColumn(resp.Values[0])
	if err != nil {
		return fmt.Errorf("failed to find name column: %v", err)
	}

	log.Printf("Found name column: %s", w.nameCol)

	// 日付列の内容を取得
	lastRow, err := w.getLastDateRow(srv, spreadsheetID, sheetName)
	if err != nil {
		return fmt.Errorf("failed to determine last date row: %v", err)
	}

	// 日付列の範囲を読み取り
	dateRange := fmt.Sprintf("%s!%s%d:%s%d", sheetName, w.dateCol, w.headerRow+1, w.dateCol, lastRow)
	dateResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, dateRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read date column: %v", err)
	}

	// 現在の日付を取得
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// シートマッパーを取得して、シート名から年月を取得
	sheetMapper, err := NewSheetMapper()
	if err != nil {
		return fmt.Errorf("failed to create sheet mapper: %v", err)
	}

	sheetMonth := sheetMapper.GetMonthFromSheetName(sheetName)
	if sheetMonth == nil {
		return fmt.Errorf("failed to determine month for sheet: %s", sheetName)
	}

	log.Printf("Sheet %s corresponds to month: %s", sheetName, sheetMonth.Format("2006-01"))

	// 更新内容を準備
	var updates []*sheets.ValueRange

	// 各日付に対して処理
	for i, row := range dateResp.Values {
		if len(row) == 0 {
			continue
		}

		// 日付が数値として存在するかチェック
		var day int
		switch v := row[0].(type) {
		case float64:
			day = int(v)
		case int:
			day = v
		case string:
			if num, err := strconv.Atoi(v); err == nil {
				day = num
			}
		default:
			continue
		}

		if day <= 0 {
			continue
		}

		// このシートのこの日の日付を計算
		cellDate := time.Date(sheetMonth.Year(), sheetMonth.Month(), day, 0, 0, 0, 0, time.Local)

		// 過去の日付はスキップ
		if cellDate.Before(today) {
			log.Printf("Skipping past date: %s (before today: %s)", cellDate.Format("2006-01-02"), today.Format("2006-01-02"))
			continue
		}

		// 該当行の行番号を計算
		rowNum := w.headerRow + i + 1

		// イベントの状態を判定
		var status string
		if events, exists := monthlyEvents[day]; exists {
			status = w.determineEventStatus(events)
		} else {
			// イベントがない日は通常勤務（渋谷）
			status = w.normalPlace
		}

		// 更新を追加
		updateRange := fmt.Sprintf("%s!%s%d", sheetName, w.nameCol, rowNum)
		updates = append(updates, &sheets.ValueRange{
			Range:  updateRange,
			Values: [][]interface{}{{status}},
		})
	}

	if len(updates) > 0 {
		log.Printf("Attempting to write %d updates to sheet %s", len(updates), sheetName)

		// バッチ更新を実行（OVERWRITE指定で既存の値を上書き）
		req := &sheets.BatchUpdateValuesRequest{
			ValueInputOption: "RAW",
			Data:             updates,
		}
		_, err = srv.Spreadsheets.Values.BatchUpdate(spreadsheetID, req).Do()
		if err != nil {
			return fmt.Errorf("failed to update values: %v", err)
		}
		log.Printf("Successfully wrote updates to sheet %s", sheetName)
	} else {
		log.Printf("No updates to write for sheet %s (all dates are in the past)", sheetName)
	}

	return nil
}
