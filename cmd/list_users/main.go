package main

import (
	"flag"
	"github.com/eotel/garoon2gs/internal/client"
	"github.com/eotel/garoon2gs/users"
	"log"
)

func main() {
	var orgID string
	flag.StringVar(&orgID, "org", "", "Organization ID to list users for")
	flag.Parse()

	garoonClient, err := client.LoadConfiguredClient()
	if err != nil {
		log.Fatal(err)
	}

	var userList []users.User
	if orgID != "" {
		userList, err = users.ListUsersByOrganization(garoonClient, orgID)
		if err != nil {
			log.Fatal("組織所属ユーザー一覧の取得に失敗しました:", err)
		}
	} else {
		userList, err = users.ListUsers(garoonClient)
		if err != nil {
			log.Fatal("ユーザー一覧の取得に失敗しました:", err)
		}
	}

	// 結果の出力
	if err := users.PrintUsers(userList); err != nil {
		log.Fatal("ユーザー一覧の出力に失敗しました:", err)
	}
}
