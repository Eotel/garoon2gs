package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/eotel/garoon2gs/internal/client"
	"github.com/joho/godotenv"
)

// Event出力用のオプション
type OutputOptions struct {
	Pretty   bool
	ShowAll  bool
	DateFrom string
	DateTo   string
}

func main() {
	// コマンドラインオプションの設定
	userID := flag.String("user", "", "ユーザーID (必須)")
	pretty := flag.Bool("pretty", false, "JSONを整形して表示する")
	showAll := flag.Bool("all", false, "すべての詳細情報を表示する")
	dateFrom := flag.String("from", "", "取得開始日 (YYYY-MM-DD形式、省略時は当月1日)")
	dateTo := flag.String("to", "", "取得終了日 (YYYY-MM-DD形式、省略時は3ヶ月後)")

	flag.Parse()

	// ユーザーIDが必須
	if *userID == "" {
		fmt.Println("Error: ユーザーIDが指定されていません")
		fmt.Println("使用法: ./get_events -user=USER_ID [オプション]")
		fmt.Println("オプション:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 設定ディレクトリの取得
	configDir, err := client.GetConfigDir()
	if err != nil {
		log.Fatal("設定ディレクトリの取得に失敗しました:", err)
	}

	// .envファイルのロード
	if err := godotenv.Load(filepath.Join(configDir, ".env")); err != nil {
		log.Println("Warning: .env ファイルが見つかりませんでした。環境変数から設定を読み込みます。")
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

	// 期間の設定
	startDate, endDate := calculateDateRange(*dateFrom, *dateTo)
	log.Printf("取得期間: %s から %s まで", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	log.Printf("ユーザーID: %s の予定を取得します", *userID)

	// 予定を取得
	events, err := garoonClient.FetchEvents(startDate, endDate, *userID)
	if err != nil {
		log.Fatal("予定の取得に失敗しました:", err)
	}

	// 結果を表示
	if len(events) == 0 {
		log.Printf("ユーザーID %s の予定は0件でした", *userID)
		return
	}

	log.Printf("ユーザーID %s の予定を %d 件取得しました", *userID, len(events))

	// 出力オプション
	opts := OutputOptions{
		Pretty:   *pretty,
		ShowAll:  *showAll,
		DateFrom: *dateFrom,
		DateTo:   *dateTo,
	}

	// イベント情報を表示
	displayEvents(events, opts)
}

// calculateDateRange は取得対象の期間を計算します
func calculateDateRange(fromStr, toStr string) (time.Time, time.Time) {
	now := time.Now()
	var startDate, endDate time.Time

	// 開始日の設定
	if fromStr != "" {
		if date, err := time.Parse("2006-01-02", fromStr); err == nil {
			startDate = date
		} else {
			log.Printf("開始日の形式が不正です: %s (YYYY-MM-DD形式で指定してください)", fromStr)
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		}
	} else {
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	}

	// 終了日の設定
	if toStr != "" {
		if date, err := time.Parse("2006-01-02", toStr); err == nil {
			// 指定された日の23:59:59までを対象にする
			endDate = date.Add(24*time.Hour - time.Second)
		} else {
			log.Printf("終了日の形式が不正です: %s (YYYY-MM-DD形式で指定してください)", toStr)
			endYear, endMonth := now.Year(), now.Month()+3
			if endMonth > 12 {
				endYear++
				endMonth = endMonth - 12
			}
			endDate = time.Date(endYear, endMonth+1, 1, 0, 0, 0, 0, time.Local).Add(-time.Second)
		}
	} else {
		endYear, endMonth := now.Year(), now.Month()+3
		if endMonth > 12 {
			endYear++
			endMonth = endMonth - 12
		}
		endDate = time.Date(endYear, endMonth+1, 1, 0, 0, 0, 0, time.Local).Add(-time.Second)
	}

	return startDate, endDate
}

// displayEvents はイベント情報を表示します
func displayEvents(events []client.Event, opts OutputOptions) {
	if opts.Pretty {
		// 整形したJSONで出力
		if opts.ShowAll {
			// すべてのフィールドを表示
			jsonData, err := json.MarshalIndent(events, "", "  ")
			if err != nil {
				log.Fatal("JSONの変換に失敗しました:", err)
			}
			fmt.Println(string(jsonData))
		} else {
			// シンプルなフィールドのみ表示
			simplifiedEvents := make([]map[string]interface{}, 0, len(events))
			for _, event := range events {
				startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
				endTime, _ := time.Parse(time.RFC3339, event.End.DateTime)

				simplifiedEvent := map[string]interface{}{
					"id":        event.ID,
					"subject":   event.Subject,
					"eventMenu": event.EventMenu,
					"start":     startTime.Format("2006-01-02 15:04"),
					"end":       endTime.Format("2006-01-02 15:04"),
				}
				simplifiedEvents = append(simplifiedEvents, simplifiedEvent)
			}

			jsonData, err := json.MarshalIndent(simplifiedEvents, "", "  ")
			if err != nil {
				log.Fatal("JSONの変換に失敗しました:", err)
			}
			fmt.Println(string(jsonData))
		}
	} else {
		// テキスト形式で出力
		for i, event := range events {
			startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
			endTime, _ := time.Parse(time.RFC3339, event.End.DateTime)

			fmt.Printf("[%d] ID: %s\n", i+1, event.ID)
			fmt.Printf("    件名: %s\n", event.Subject)
			if event.EventMenu != "" {
				fmt.Printf("    予定メニュー: %s\n", event.EventMenu)
			}
			fmt.Printf("    期間: %s - %s\n",
				startTime.Format("2006-01-02 15:04"),
				endTime.Format("2006-01-02 15:04"))

			if i < len(events)-1 {
				fmt.Println("---")
			}
		}
	}
}
