package organizations

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListOrganizationsPaginatesAllOrganizations(t *testing.T) {
	t.Helper()

	var offsets []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsets = append(offsets, r.URL.Query().Get("offset"))
		if got := r.URL.Query().Get("limit"); got != "1000" {
			t.Fatalf("limit = %s, want 1000", got)
		}

		var response listOrganizationsResponse
		switch r.URL.Query().Get("offset") {
		case "0":
			response = listOrganizationsResponse{
				Organizations: []Organization{
					{ID: "1", Name: "営業部"},
					{ID: "2", Name: "開発部"},
				},
				HasNext: true,
			}
		case "2":
			response = listOrganizationsResponse{
				Organizations: []Organization{
					{ID: "3", Name: "管理部"},
				},
				HasNext: false,
			}
		default:
			t.Fatalf("unexpected offset: %s", r.URL.Query().Get("offset"))
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	orgs, err := ListOrganizations(server.Client(), server.URL, "user", "pass")
	if err != nil {
		t.Fatalf("ListOrganizations returned error: %v", err)
	}

	if len(orgs) != 3 {
		t.Fatalf("len(orgs) = %d, want 3", len(orgs))
	}

	if len(offsets) != 2 || offsets[0] != "0" || offsets[1] != "2" {
		t.Fatalf("offsets = %v, want [0 2]", offsets)
	}
}
