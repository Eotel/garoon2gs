package organizations

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Organization represents a Garoon organization
type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code,omitempty"`
	ParentID    string `json:"parentId,omitempty"`
	Description string `json:"description,omitempty"`
}

type listOrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
	HasNext       bool           `json:"hasNext"`
}

const maxPageLimit = 1000

// ListOrganizations retrieves all organizations
func ListOrganizations(client *http.Client, baseURL, username, password string) ([]Organization, error) {
	var allOrganizations []Organization
	offset := 0

	for {
		orgs, hasNext, err := fetchOrganizationsPage(client, baseURL, username, password, offset, maxPageLimit)
		if err != nil {
			return nil, err
		}

		allOrganizations = append(allOrganizations, orgs...)
		if !hasNext {
			return allOrganizations, nil
		}

		offset += len(orgs)
		if len(orgs) == 0 {
			return nil, fmt.Errorf("API returned hasNext=true but no organizations at offset=%d", offset)
		}
	}
}

func fetchOrganizationsPage(client *http.Client, baseURL, username, password string, offset, limit int) ([]Organization, bool, error) {
	params := url.Values{}
	params.Set("offset", strconv.Itoa(offset))
	params.Set("limit", strconv.Itoa(limit))

	reqURL := fmt.Sprintf("%s/api/v1/base/organizations?%s", baseURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("リクエストの作成に失敗しました: %v", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	req.Header.Set("X-Cybozu-Authorization", auth)

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("APIリクエストに失敗しました: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("APIエラー（ステータスコード: %d）: %s", resp.StatusCode, string(body))
	}

	var response listOrganizationsResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, false, fmt.Errorf("JSONのデコードに失敗しました: %v", err)
	}

	return response.Organizations, response.HasNext, nil
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
