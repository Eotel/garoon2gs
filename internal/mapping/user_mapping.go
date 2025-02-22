package mapping

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// UserMapping はユーザーIDと列名のマッピングを表す構造体です
type UserMapping struct {
	UserID     string // Garoonのユーザーid
	HeaderName string // スプレッドシートのヘッダーに表示される名前
}

// LoadUserMapping はCSVファイルからユーザーマッピングを読み込みます
func LoadUserMapping(configDir string) ([]UserMapping, error) {
	// 環境変数からCSVファイルのパスを取得
	csvPathFromEnv := os.Getenv("USER_MAPPING_PATH")
	if csvPathFromEnv == "" {
		return nil, fmt.Errorf("USER_MAPPING_PATH environment variable is not set")
	}

	// CSVファイルの絶対パスを構築
	csvPath := filepath.Join(configDir, csvPathFromEnv)

	// CSVファイルを開く
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open user mapping CSV file %s: %v", csvPath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// ヘッダーを読み飛ばす
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	// ヘッダーの検証
	if len(header) != 2 || header[0] != "user_id" || header[1] != "name" {
		return nil, fmt.Errorf("invalid CSV header format: expected [user_id,name] but got %v", header)
	}

	var mappings []UserMapping

	// 各行を読み込む
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %v", err)
		}

		if len(record) != 2 {
			return nil, fmt.Errorf("invalid CSV format: expected 2 columns but got %d", len(record))
		}

		// UserMappingを作成（name_colをHeaderNameとして保存）
		mappings = append(mappings, UserMapping{
			UserID:     record[0],
			HeaderName: record[1],
		})
	}

	log.Printf("Loaded %d user mappings from %s", len(mappings), csvPath)
	return mappings, nil
}

// GetColumnForUser は指定されたユーザーIDに対応するヘッダー名を返します
func GetColumnForUser(mappings []UserMapping, userID string) (string, bool) {
	for _, m := range mappings {
		if m.UserID == userID {
			return m.HeaderName, true
		}
	}
	return "", false
}
