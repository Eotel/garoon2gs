package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// UserMapping はユーザーIDと列名のマッピングを表す構造体です
type UserMapping struct {
	UserID  string
	NameCol string
}

// UserMapper はユーザーIDと列名を解決するマッパーです
type UserMapper struct {
	mappings []UserMapping
}

// NewUserMapper は新しいUserMapperインスタンスを作成します
func NewUserMapper() (*UserMapper, error) {
	// 設定ディレクトリの取得
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %v", err)
	}

	// 環境変数からCSVファイルのパスを取得
	csvPathFromEnv := os.Getenv("USER_MAPPING_PATH")
	if csvPathFromEnv == "" {
		return nil, fmt.Errorf("USER_MAPPING_PATH environment variable is not set")
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

	var mappings []UserMapping

	// 各行を読み込む
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %v", err)
	}

	for _, record := range records {
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid CSV format: expected 2 columns but got %d", len(record))
		}

		mappings = append(mappings, UserMapping{
			UserID:  record[0],
			NameCol: record[1],
		})
	}

	log.Printf("Loaded %d user mappings from %s", len(mappings), csvPath)
	return &UserMapper{
		mappings: mappings,
	}, nil
}

// GetAllMappings はすべてのマッピングを返します
func (um *UserMapper) GetAllMappings() []UserMapping {
	return um.mappings
}
