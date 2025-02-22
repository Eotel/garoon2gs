package main

import (
	"github.com/eotel/garoon2gs/internal/client"
	"github.com/eotel/garoon2gs/users"
	"log"
)

func main() {
	// 設定の読み込みとクライアントの初期化
	config, err := client.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	garoonClient, err := client.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// ユーザー一覧の取得
	userList, err := users.ListUsers(
		garoonClient.GetHTTPClient(),
		garoonClient.GetBaseURL(),
		garoonClient.GetUsername(),
		garoonClient.GetPassword(),
	)
	if err != nil {
		log.Fatal("ユーザー一覧の取得に失敗しました:", err)
	}

	// 結果の出力
	if err := users.PrintUsers(userList); err != nil {
		log.Fatal("ユーザー一覧の出力に失敗しました:", err)
	}
}
