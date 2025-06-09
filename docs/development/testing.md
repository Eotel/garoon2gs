# テストガイド

Garoon2GSのテストの書き方と実行方法について説明します。

## テストの原則

1. **Table-Driven Tests**: 複数のテストケースを効率的に管理
2. **明確なテスト名**: 日本語での説明的な名前を使用
3. **独立性**: 各テストは他のテストに依存しない
4. **モックの活用**: 外部依存をモック化

## テストの構成

### ディレクトリ構成

```
garoon2gs/
├── garoon2gs_test.go          # メイン関数のテスト
├── schedule_writer_test.go     # スケジュール書き込みテスト
├── sheet_mapper_test.go        # シートマッピングテスト
├── internal/
│   ├── client/
│   │   └── client_test.go     # APIクライアントテスト
│   └── mapping/
│       └── user_mapping_test.go # マッピングテスト
└── testdata/                   # テスト用データファイル
    ├── user_mapping.csv
    └── sheet_mapping.csv
```

## テストの書き方

### 1. Table-Driven Testパターン

```go
func TestScheduleWriter_DetermineStatus(t *testing.T) {
    testCases := []struct {
        name     string
        events   []Event
        expected string
    }{
        {
            name: "休暇イベントがある場合",
            events: []Event{
                {EventMenu: "年次休暇"},
            },
            expected: "休み",
        },
        {
            name: "外出イベントがある場合",
            events: []Event{
                {EventMenu: "外出"},
            },
            expected: "外出",
        },
        {
            name: "イベントがない場合は通常勤務",
            events:   []Event{},
            expected: "渋谷",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            writer := &ScheduleWriter{
                holidayMenus: []string{"年次休暇", "週休"},
                outingMenus:  []string{"外出", "出張"},
                normalPlace:  "渋谷",
            }
            
            result := writer.determineStatus(tc.events)
            if result != tc.expected {
                t.Errorf("expected %s, got %s", tc.expected, result)
            }
        })
    }
}
```

### 2. 環境変数のモック

```go
func TestLoadConfig(t *testing.T) {
    // 環境変数を一時的に設定
    t.Setenv("GAROON_BASE_URL", "https://test.cybozu.com/g")
    t.Setenv("GAROON_USERNAME", "testuser")
    t.Setenv("GAROON_PASSWORD", "testpass")
    
    cfg, err := LoadConfig()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if cfg.GaroonBaseURL != "https://test.cybozu.com/g" {
        t.Errorf("expected base URL to be set")
    }
}
```

### 3. HTTPクライアントのモック

```go
type mockHTTPClient struct {
    responses map[string]*http.Response
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    resp, ok := m.responses[req.URL.Path]
    if !ok {
        return nil, fmt.Errorf("unexpected request: %s", req.URL.Path)
    }
    return resp, nil
}

func TestClient_FetchEvents(t *testing.T) {
    mockResp := &http.Response{
        StatusCode: 200,
        Body: io.NopCloser(strings.NewReader(`{
            "events": [
                {"id": "1", "subject": "会議"}
            ]
        }`)),
    }
    
    client := &Client{
        httpClient: &mockHTTPClient{
            responses: map[string]*http.Response{
                "/api/v1/schedule/events": mockResp,
            },
        },
    }
    
    events, err := client.FetchEvents(userID, start, end)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if len(events) != 1 {
        t.Errorf("expected 1 event, got %d", len(events))
    }
}
```

### 4. ファイル操作のテスト

```go
func TestLoadUserMapping(t *testing.T) {
    // テスト用の一時ファイルを作成
    tmpFile := filepath.Join(t.TempDir(), "user_mapping.csv")
    content := `12345,山田太郎
67890,田中花子`
    
    if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
        t.Fatalf("failed to create test file: %v", err)
    }
    
    mapping, err := LoadUserMapping(tmpFile)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if mapping["12345"] != "山田太郎" {
        t.Errorf("expected '山田太郎', got '%s'", mapping["12345"])
    }
}
```

### 5. エラーケースのテスト

```go
func TestScheduleWriter_WriteSchedule_Error(t *testing.T) {
    testCases := []struct {
        name        string
        setupFunc   func() *ScheduleWriter
        expectError string
    }{
        {
            name: "シートが見つからない場合",
            setupFunc: func() *ScheduleWriter {
                return &ScheduleWriter{
                    sheetsService: mockSheetsServiceWithError("sheet not found"),
                }
            },
            expectError: "sheet not found",
        },
        {
            name: "権限エラーの場合",
            setupFunc: func() *ScheduleWriter {
                return &ScheduleWriter{
                    sheetsService: mockSheetsServiceWithError("permission denied"),
                }
            },
            expectError: "permission denied",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            writer := tc.setupFunc()
            err := writer.WriteSchedule(date, userID, events)
            
            if err == nil {
                t.Fatal("expected error but got nil")
            }
            
            if !strings.Contains(err.Error(), tc.expectError) {
                t.Errorf("expected error containing '%s', got '%v'", 
                    tc.expectError, err)
            }
        })
    }
}
```

## テストヘルパー関数

### テストデータの作成

```go
// test_helpers.go
package main

import (
    "testing"
    "time"
)

func createTestEvent(menu string, date time.Time) Event {
    return Event{
        ID:        "test-1",
        Subject:   "Test Event",
        EventMenu: menu,
        Start:     EventTime{DateTime: date},
        End:       EventTime{DateTime: date.Add(time.Hour)},
    }
}

func createTestScheduleWriter(t *testing.T) *ScheduleWriter {
    t.Helper()
    
    return &ScheduleWriter{
        spreadsheetID: "test-sheet-id",
        holidayMenus:  []string{"休み", "年次休暇"},
        outingMenus:   []string{"外出", "出張"},
        normalPlace:   "渋谷",
        headerRow:     7,
        dateCol:       "A",
    }
}
```

## テストの実行

### 基本的なテスト実行

```bash
# 全テストの実行
go test ./...

# 詳細な出力
go test -v ./...

# 特定のパッケージのテスト
go test ./internal/client

# 特定のテスト関数の実行
go test -v -run TestScheduleWriter_DetermineStatus

# 並列実行
go test -parallel 4 ./...
```

### カバレッジの測定

```bash
# カバレッジレポートの生成
go test -coverprofile=coverage.out ./...

# HTMLレポートの表示
go tool cover -html=coverage.out

# カバレッジの概要
go test -cover ./...
```

### ベンチマークテスト

```go
func BenchmarkDetermineStatus(b *testing.B) {
    writer := &ScheduleWriter{
        holidayMenus: []string{"休み", "年次休暇"},
        outingMenus:  []string{"外出", "出張"},
        normalPlace:  "渋谷",
    }
    
    events := []Event{
        {EventMenu: "会議"},
        {EventMenu: "作業"},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = writer.determineStatus(events)
    }
}
```

実行：
```bash
go test -bench=. -benchmem
```

## 統合テスト

### 実際のAPIを使用したテスト

```go
// +build integration

func TestIntegration_FullSync(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    // 実際の環境変数を使用
    cfg, err := LoadConfig()
    if err != nil {
        t.Fatalf("failed to load config: %v", err)
    }
    
    // 実際のAPIを呼び出してテスト
    // ...
}
```

実行：
```bash
# 統合テストを含む実行
go test -tags=integration ./...

# 統合テストを除外
go test -short ./...
```

## テストのベストプラクティス

### 1. 明確なテスト名

```go
// Good
t.Run("祝日と外出が同じ日にある場合は祝日を優先", func(t *testing.T) {
    // ...
})

// Bad
t.Run("test1", func(t *testing.T) {
    // ...
})
```

### 2. Arrange-Act-Assert パターン

```go
func TestSomething(t *testing.T) {
    // Arrange - テストの準備
    writer := createTestScheduleWriter(t)
    events := []Event{createTestEvent("会議", time.Now())}
    
    // Act - テスト対象の実行
    result := writer.determineStatus(events)
    
    // Assert - 結果の検証
    if result != "渋谷" {
        t.Errorf("expected '渋谷', got '%s'", result)
    }
}
```

### 3. テストの独立性

```go
// 各テストで新しいインスタンスを作成
func TestIndependent(t *testing.T) {
    t.Run("test1", func(t *testing.T) {
        writer := createTestScheduleWriter(t) // 新しいインスタンス
        // ...
    })
    
    t.Run("test2", func(t *testing.T) {
        writer := createTestScheduleWriter(t) // 新しいインスタンス
        // ...
    })
}
```

## CI/CDでのテスト

### GitHub Actions設定例

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: |
        go test -v -cover ./...
        
    - name: Run lint
      run: |
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        golangci-lint run
```

## トラブルシューティング

### テストが失敗する場合

1. **環境依存の問題**
   - `t.Setenv()`を使用して環境変数を設定
   - 一時ディレクトリ`t.TempDir()`を使用

2. **並行実行の問題**
   - `t.Parallel()`の使用を確認
   - 共有リソースへのアクセスを確認

3. **タイムゾーンの問題**
   - 明示的にタイムゾーンを指定
   - `time.Local`の使用を避ける