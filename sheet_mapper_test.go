package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestEnvironment(t *testing.T) (func(), error) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()

	// テスト用のCSVファイルを作成
	csvContent := `month,sheet_name
2025-02,R6年度_2月
2025-03,R6年度_3月
2025-04,R7年度_4月`

	csvPath := filepath.Join(tmpDir, "sheet_mapping.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		return nil, err
	}

	// 既存の環境変数とカレントディレクトリを保存
	originalPath := os.Getenv("SHEET_MAPPING_PATH")
	originalWd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// 環境変数を直接設定
	os.Setenv("SHEET_MAPPING_PATH", csvPath)

	// テストディレクトリに移動
	if err := os.Chdir(tmpDir); err != nil {
		return nil, err
	}

	// クリーンアップ関数を返す
	cleanup := func() {
		os.Setenv("SHEET_MAPPING_PATH", originalPath)
		os.Chdir(originalWd)
	}

	return cleanup, nil
}

func TestSheetMapper(t *testing.T) {
	cleanup, err := setupTestEnvironment(t)
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}
	defer cleanup()

	// SheetMapperの初期化
	mapper, err := NewSheetMapper()
	if err != nil {
		t.Fatalf("Failed to create SheetMapper: %v", err)
	}

	tests := []struct {
		name          string
		eventDate     time.Time
		expectedSheet *string
	}{
		{
			name:          "2025年2月のイベント",
			eventDate:     time.Date(2025, 2, 15, 0, 0, 0, 0, time.Local),
			expectedSheet: strPtr("R6年度_2月"),
		},
		{
			name:          "2025年3月のイベント",
			eventDate:     time.Date(2025, 3, 1, 0, 0, 0, 0, time.Local),
			expectedSheet: strPtr("R6年度_3月"),
		},
		{
			name:          "2025年4月のイベント",
			eventDate:     time.Date(2025, 4, 30, 0, 0, 0, 0, time.Local),
			expectedSheet: strPtr("R7年度_4月"),
		},
		{
			name:          "マッピングにない月のイベント",
			eventDate:     time.Date(2025, 5, 1, 0, 0, 0, 0, time.Local),
			expectedSheet: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sheet := mapper.GetSheetName(tt.eventDate)

			if (sheet == nil) != (tt.expectedSheet == nil) {
				t.Errorf("expected sheet name %v but got %v", tt.expectedSheet, sheet)
			}
			if sheet != nil && tt.expectedSheet != nil && *sheet != *tt.expectedSheet {
				t.Errorf("expected sheet name %q but got %q", *tt.expectedSheet, *sheet)
			}
		})
	}
}

func TestSheetMapperWithInvalidCSV(t *testing.T) {
	tmpDir := t.TempDir()

	// 不正なCSVファイルを作成
	invalidCSV := `invalid,format
2025-02`

	csvPath := filepath.Join(tmpDir, "invalid_mapping.csv")
	if err := os.WriteFile(csvPath, []byte(invalidCSV), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	// 環境変数を設定
	originalPath := os.Getenv("SHEET_MAPPING_PATH")
	os.Setenv("SHEET_MAPPING_PATH", csvPath)
	defer os.Setenv("SHEET_MAPPING_PATH", originalPath)

	// 不正なCSVでの初期化テスト
	_, err := NewSheetMapper()
	if err == nil {
		t.Error("expected error with invalid CSV but got none")
	}
}

func TestSheetMapperWithNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 存在しないファイルのパスを環境変数に設定
	nonExistentPath := filepath.Join(tmpDir, "non_existent.csv")
	originalPath := os.Getenv("SHEET_MAPPING_PATH")
	os.Setenv("SHEET_MAPPING_PATH", nonExistentPath)
	defer os.Setenv("SHEET_MAPPING_PATH", originalPath)

	// 存在しないファイルでの初期化テスト
	_, err := NewSheetMapper()
	if err == nil {
		t.Error("expected error with non-existent file but got none")
	}
}

// TestSheetMapperWithoutEnvVar は環境変数が設定されていない場合のテスト
func TestSheetMapperWithoutEnvVar(t *testing.T) {
	// 環境変数をクリア
	originalPath := os.Getenv("SHEET_MAPPING_PATH")
	os.Unsetenv("SHEET_MAPPING_PATH")
	defer os.Setenv("SHEET_MAPPING_PATH", originalPath)

	// 初期化テスト
	_, err := NewSheetMapper()
	if err == nil {
		t.Error("expected error when SHEET_MAPPING_PATH is not set")
	}
}

// strPtr は文字列のポインタを返すヘルパー関数です
func strPtr(s string) *string {
	return &s
}
