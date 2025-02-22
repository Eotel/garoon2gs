package main

import (
	"flag"
	"github.com/eotel/garoon2gs/internal/client"
	"github.com/eotel/garoon2gs/organizations"
	"github.com/eotel/garoon2gs/users"
	"log"
)

func main() {
	var orgID string
	flag.StringVar(&orgID, "org", "", "Organization ID to list users for")
	flag.Parse()

	// 設定の読み込みとクライアントの初期化
	config, err := client.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	garoonClient, err := client.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// 組織IDが指定された場合は組織メンバーを表示
	if orgID != "" {
		userList, err := organizations.GetOrganizationUsers(
			garoonClient.GetHTTPClient(),
			garoonClient.GetBaseURL(),
			garoonClient.GetUsername(),
			garoonClient.GetPassword(),
			orgID,
		)
		if err != nil {
			log.Fatalf("組織メンバーの取得に失敗しました: %v", err)
		}
		if err := users.PrintUsers(userList); err != nil {
			log.Fatal("ユーザー一覧の出力に失敗しました:", err)
		}
		return
	}

	// 組織IDが指定されていない場合は組織一覧を表示
	orgs, err := organizations.ListOrganizations(
		garoonClient.GetHTTPClient(),
		garoonClient.GetBaseURL(),
		garoonClient.GetUsername(),
		garoonClient.GetPassword(),
	)
	if err != nil {
		log.Fatal("組織一覧の取得に失敗しました:", err)
	}

	if err := organizations.PrintOrganizations(orgs); err != nil {
		log.Fatal("組織一覧の出力に失敗しました:", err)
	}
}
