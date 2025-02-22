package organizations

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/eotel/garoon2gs/users"
	"io"
	"net/http"
)

// Organization represents a Garoon organization
type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code,omitempty"`
	ParentID    string `json:"parentId,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListOrganizations retrieves all organizations
func ListOrganizations(client *http.Client, baseURL, username, password string) ([]Organization, error) {
	reqURL := fmt.Sprintf("%s/api/v1/base/organizations", baseURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %v", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	req.Header.Set("X-Cybozu-Authorization", auth)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストに失敗しました: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("APIエラー（ステータスコード: %d）: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Organizations []Organization `json:"organizations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("JSONのデコードに失敗しました: %v", err)
	}

	return response.Organizations, nil
}

// GetOrganizationUsers retrieves users belonging to a specific organization
func GetOrganizationUsers(client *http.Client, baseURL, username, password, orgID string) ([]users.User, error) {
	reqURL := fmt.Sprintf("%s/api/v1/base/organizations/%s/users", baseURL, orgID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %v", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	req.Header.Set("X-Cybozu-Authorization", auth)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストに失敗しました: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("APIエラー（ステータスコード: %d）: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Users []users.User `json:"users"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("JSONのデコードに失敗しました: %v", err)
	}

	return response.Users, nil
}

// PrintOrganizations formats and prints organization list
func PrintOrganizations(orgs []Organization) error {
	prettyJSON, err := json.MarshalIndent(struct {
		Organizations []Organization `json:"organizations"`
	}{orgs}, "", "  ")
	if err != nil {
		return fmt.Errorf("JSONの整形に失敗しました: %v", err)
	}

	fmt.Println(string(prettyJSON))
	return nil
}
