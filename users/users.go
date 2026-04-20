package users

import (
	"encoding/json"
	"fmt"
	"github.com/eotel/garoon2gs/internal/client"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// User represents a Garoon user
type User struct {
	ID                  string `json:"id"`
	Code                string `json:"code"`
	Name                string `json:"name"`
	Email               string `json:"email,omitempty"`
	Status              string `json:"status"`
	PrimaryOrganization struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"primaryOrganization"`
}

const maxPageLimit = 1000

type listUsersResponse struct {
	Users   []User `json:"users"`
	HasNext bool   `json:"hasNext"`
}

// ListUsers はユーザー一覧を取得する関数です
func ListUsers(garoonClient *client.GaroonClient) ([]User, error) {
	return listUsers(garoonClient, "")
}

// ListUsersByOrganization は指定した組織に所属するユーザー一覧を取得する関数です
func ListUsersByOrganization(garoonClient *client.GaroonClient, orgID string) ([]User, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	return listUsers(garoonClient, fmt.Sprintf("/api/v1/base/organizations/%s/users", orgID))
}

func listUsers(garoonClient *client.GaroonClient, path string) ([]User, error) {
	if path == "" {
		path = "/api/v1/base/users"
	}

	var allUsers []User
	offset := 0

	for {
		users, hasNext, err := fetchUsersPage(garoonClient, path, offset, maxPageLimit)
		if err != nil {
			return nil, err
		}

		allUsers = append(allUsers, users...)
		if !hasNext {
			return allUsers, nil
		}

		offset += len(users)
		if len(users) == 0 {
			return nil, fmt.Errorf("API returned hasNext=true but no users at offset=%d", offset)
		}
	}
}

func fetchUsersPage(garoonClient *client.GaroonClient, path string, offset, limit int) ([]User, bool, error) {
	params := url.Values{}
	params.Set("offset", strconv.Itoa(offset))
	params.Set("limit", strconv.Itoa(limit))

	reqURL := fmt.Sprintf("%s%s?%s", garoonClient.GetBaseURL(), path, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("リクエストの作成に失敗しました: %v", err)
	}

	garoonClient.ApplyAuth(req)

	// リクエストの実行
	resp, err := garoonClient.GetHTTPClient().Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("APIリクエストに失敗しました: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスのステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("APIエラー（ステータスコード: %d）: %s", resp.StatusCode, string(body))
	}

	// レスポンスの読み取りとJSON解析
	var response listUsersResponse

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, false, fmt.Errorf("JSONのデコードに失敗しました: %v", err)
	}

	return response.Users, response.HasNext, nil
}

// PrintUsers はユーザー一覧を整形して出力する関数です
func PrintUsers(users []User) error {
	// 結果の整形と出力
	prettyJSON, err := json.MarshalIndent(struct {
		Users []User `json:"users"`
	}{users}, "", "  ")
	if err != nil {
		return fmt.Errorf("JSONの整形に失敗しました: %v", err)
	}

	fmt.Println(string(prettyJSON))
	return nil
}
