package main

import (
	"github.com/eotel/garoon2gs/internal/client"
	"os"
	"testing"
	"time"
)

// Event はテストで使用するイベント構造体です
type Event = client.Event

func setupWriterTest(t *testing.T) func() {
	// テスト用の環境変数を設定
	originalEnv := map[string]string{}
	envVars := map[string]string{
		"HEADER_ROW": "7",
		"DATE_COL":   "A",
		"NAME":       "伊藤",
	}

	// 既存の環境変数を保存し、テスト用の値を設定
	for k, v := range envVars {
		if original, exists := os.LookupEnv(k); exists {
			originalEnv[k] = original
		}
		os.Setenv(k, v)
	}

	// クリーンアップ関数を返す
	return func() {
		for k := range envVars {
			if original, exists := originalEnv[k]; exists {
				os.Setenv(k, original)
			} else {
				os.Unsetenv(k)
			}
		}
	}
}

func TestScheduleWriter(t *testing.T) {
	cleanup := setupWriterTest(t)
	defer cleanup()

	writer, err := NewScheduleWriter()
	if err != nil {
		t.Fatalf("Failed to create ScheduleWriter: %v", err)
	}
	
	writer.name = "伊藤"

	// テストのために名前の列を設定
	writer.nameCol = "J"

	// getCellPosition のテスト
	tests := []struct {
		name           string
		date           time.Time
		expectedRow    int
		expectedColumn string
		expectError    bool
	}{
		{
			name:           "2月1日の位置",
			date:           time.Date(2025, 2, 1, 0, 0, 0, 0, time.Local),
			expectedRow:    8,   // HEADER_ROW + 1
			expectedColumn: "J", // "伊藤"の列
			expectError:    false,
		},
		{
			name:           "2月15日の位置",
			date:           time.Date(2025, 2, 15, 0, 0, 0, 0, time.Local),
			expectedRow:    22, // HEADER_ROW + 15
			expectedColumn: "J",
			expectError:    false,
		},
		{
			name:           "2月28日の位置",
			date:           time.Date(2025, 2, 28, 0, 0, 0, 0, time.Local),
			expectedRow:    35, // HEADER_ROW + 28
			expectedColumn: "J",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, col, err := writer.getCellPosition(tt.date)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if row != tt.expectedRow {
				t.Errorf("expected row %d but got %d", tt.expectedRow, row)
			}

			if col != tt.expectedColumn {
				t.Errorf("expected column %s but got %s", tt.expectedColumn, col)
			}
		})
	}
}

func TestDetermineEventStatus(t *testing.T) {
	cleanup := setupWriterTest(t)
	defer cleanup()

	writer, err := NewScheduleWriter()
	if err != nil {
		t.Fatalf("Failed to create ScheduleWriter: %v", err)
	}
	
	writer.name = "伊藤"

	// holidayMenusを設定
	writer.holidayMenus = []string{"休暇", "週休"}
	// normalPlaceを設定
	writer.normalPlace = "渋谷"

	tests := []struct {
		name          string
		events        []Event
		expectedValue string
	}{
		{
			name: "休暇がある場合は週休",
			events: []Event{
				{EventMenu: "イベント"},
				{EventMenu: "休暇"},
				{EventMenu: "ミーティング"},
			},
			expectedValue: "週休",
		},
		{
			name: "週休がある場合は週休",
			events: []Event{
				{EventMenu: "イベント"},
				{EventMenu: "週休"},
				{EventMenu: "ミーティング"},
			},
			expectedValue: "週休",
		},
		{
			name: "休暇も週休もない場合は渋谷",
			events: []Event{
				{EventMenu: "イベント"},
				{EventMenu: "ミーティング"},
			},
			expectedValue: "渋谷",
		},
		{
			name:          "イベントがない場合は渋谷",
			events:        []Event{},
			expectedValue: "渋谷",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.determineEventStatus(tt.events)
			if result != tt.expectedValue {
				t.Errorf("expected %s but got %s", tt.expectedValue, result)
			}
		})
	}
}

func TestFindNameColumn(t *testing.T) {
	cleanup := setupWriterTest(t)
	defer cleanup()

	writer, err := NewScheduleWriter()
	if err != nil {
		t.Fatalf("Failed to create ScheduleWriter: %v", err)
	}
	
	writer.name = "伊藤"

	tests := []struct {
		name         string
		headerValues []interface{}
		expectedCol  string
		expectError  bool
	}{
		{
			name:         "名前が見つかる場合",
			headerValues: []interface{}{"DATE", "DoW", "予定", "伊藤"},
			expectedCol:  "D",
			expectError:  false,
		},
		{
			name:         "名前が見つからない場合",
			headerValues: []interface{}{"DATE", "DoW", "予定", "山田"},
			expectedCol:  "",
			expectError:  true,
		},
		{
			name:         "空のヘッダー",
			headerValues: []interface{}{},
			expectedCol:  "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, err := writer.findNameColumn(tt.headerValues)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if col != tt.expectedCol {
				t.Errorf("expected column %s but got %s", tt.expectedCol, col)
			}
		})
	}
}
