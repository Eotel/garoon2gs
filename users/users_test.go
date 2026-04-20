package users

import (
	"encoding/json"
	"github.com/eotel/garoon2gs/internal/client"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestGaroonClient(t *testing.T, baseURL string, authType client.AuthType) *client.GaroonClient {
	t.Helper()

	cfg := &client.Config{BaseURL: baseURL}
	switch authType {
	case client.AuthTypeOAuth:
		cfg.AuthType = client.AuthTypeOAuth
		cfg.BearerToken = "token"
	default:
		cfg.Username = "user"
		cfg.Password = "pass"
	}

	garoonClient, err := client.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	return garoonClient
}

func TestListUsersPaginatesAllUsers(t *testing.T) {
	t.Helper()

	var offsets []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsets = append(offsets, r.URL.Query().Get("offset"))
		if got := r.URL.Query().Get("limit"); got != "1000" {
			t.Fatalf("limit = %s, want 1000", got)
		}

		var response listUsersResponse
		switch r.URL.Query().Get("offset") {
		case "0":
			response = listUsersResponse{
				Users: []User{
					{ID: "1", Name: "伊藤"},
					{ID: "2", Name: "島田"},
				},
				HasNext: true,
			}
		case "2":
			response = listUsersResponse{
				Users: []User{
					{ID: "3", Name: "稲田"},
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

	users, err := ListUsers(newTestGaroonClient(t, server.URL, client.AuthTypePassword))
	if err != nil {
		t.Fatalf("ListUsers returned error: %v", err)
	}

	if len(users) != 3 {
		t.Fatalf("len(users) = %d, want 3", len(users))
	}

	if len(offsets) != 2 || offsets[0] != "0" || offsets[1] != "2" {
		t.Fatalf("offsets = %v, want [0 2]", offsets)
	}
}

func TestListUsersByOrganizationUsesOrganizationEndpoint(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/api/v1/base/organizations/42/users"; got != want {
			t.Fatalf("path = %s, want %s", got, want)
		}

		if err := json.NewEncoder(w).Encode(listUsersResponse{
			Users: []User{
				{ID: "10", Name: "原"},
			},
			HasNext: false,
		}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	users, err := ListUsersByOrganization(newTestGaroonClient(t, server.URL, client.AuthTypePassword), "42")
	if err != nil {
		t.Fatalf("ListUsersByOrganization returned error: %v", err)
	}

	if len(users) != 1 || users[0].ID != "10" {
		t.Fatalf("users = %+v, want one user with ID 10", users)
	}
}

func TestListUsersUsesBearerTokenForOAuth(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer token")
		}
		if got := r.Header.Get("X-Cybozu-Authorization"); got != "" {
			t.Fatalf("X-Cybozu-Authorization = %q, want empty", got)
		}

		if err := json.NewEncoder(w).Encode(listUsersResponse{HasNext: false}); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	if _, err := ListUsers(newTestGaroonClient(t, server.URL, client.AuthTypeOAuth)); err != nil {
		t.Fatalf("ListUsers returned error: %v", err)
	}
}
