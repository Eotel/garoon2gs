package main

import (
	"github.com/eotel/garoon2gs/internal/client"
	"github.com/eotel/garoon2gs/organizations"
	"log"
)

func main() {
	garoonClient, err := client.LoadConfiguredClient()
	if err != nil {
		log.Fatal(err)
	}

	orgs, err := organizations.ListOrganizations(garoonClient)
	if err != nil {
		log.Fatal("組織一覧の取得に失敗しました:", err)
	}

	if err := organizations.PrintOrganizations(orgs); err != nil {
		log.Fatal("組織一覧の出力に失敗しました:", err)
	}
}
