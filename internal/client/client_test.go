package client

import (
	"net/http"
	"testing"
)

func TestConfigValidateRequiresPasswordAuth(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:  "https://example.cybozu.com/g",
		Username: "",
		Password: "",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() returned nil, want error")
	}
}

func TestConfigValidateRequiresBearerTokenForOAuth(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:  "https://example.cybozu.com/g",
		AuthType: AuthTypeOAuth,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() returned nil, want error")
	}
}

func TestConfigValidateRejectsUnknownAuthType(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:  "https://example.cybozu.com/g",
		AuthType: AuthType("unknown"),
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() returned nil, want error")
	}
}

func TestConfigValidateRequiresClientCertPair(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:      "https://example.cybozu.com/g",
		Username:     "user",
		Password:     "pass",
		CertPath:     "/tmp/client.pfx",
		CertPassword: "",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() returned nil, want error")
	}
}

func TestConfigValidateAllowsOptionalClientCertificate(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:      "https://example.cybozu.com/g",
		Username:     "user",
		Password:     "pass",
		CertPath:     "/tmp/client.pfx",
		CertPassword: "secret",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
}

func TestConfigValidateAllowsOAuthWithoutPassword(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:     "https://example.cybozu.com/g",
		AuthType:    AuthTypeOAuth,
		BearerToken: "token",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
}

func TestLoadConfiguredClientSuccessWithoutClientCertificate(t *testing.T) {
	t.Helper()

	t.Setenv("GAROON_BASE_URL", "https://example.cybozu.com/g")
	t.Setenv("GAROON_USERNAME", "user")
	t.Setenv("GAROON_PASSWORD", "pass")
	t.Setenv("CLIENT_CERT_PATH", "")
	t.Setenv("CLIENT_CERT_PASSWORD", "")

	client, err := LoadConfiguredClient()
	if err != nil {
		t.Fatalf("LoadConfiguredClient() returned error: %v", err)
	}

	if client.GetBaseURL() != "https://example.cybozu.com/g" {
		t.Fatalf("GetBaseURL() = %s", client.GetBaseURL())
	}
}

func TestLoadConfiguredClientSuccessWithOAuth(t *testing.T) {
	t.Helper()

	t.Setenv("GAROON_BASE_URL", "https://example.cybozu.com/g")
	t.Setenv("GAROON_AUTH_TYPE", "oauth")
	t.Setenv("GAROON_BEARER_TOKEN", "token")
	t.Setenv("GAROON_USERNAME", "")
	t.Setenv("GAROON_PASSWORD", "")
	t.Setenv("CLIENT_CERT_PATH", "")
	t.Setenv("CLIENT_CERT_PASSWORD", "")

	client, err := LoadConfiguredClient()
	if err != nil {
		t.Fatalf("LoadConfiguredClient() returned error: %v", err)
	}

	req, err := http.NewRequest("GET", client.GetBaseURL(), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}

	client.ApplyAuth(req)
	if got := req.Header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer token")
	}
}

func TestLoadConfiguredClientRejectsIncompleteClientCertificate(t *testing.T) {
	t.Helper()

	t.Setenv("GAROON_BASE_URL", "https://example.cybozu.com/g")
	t.Setenv("GAROON_USERNAME", "user")
	t.Setenv("GAROON_PASSWORD", "pass")
	t.Setenv("CLIENT_CERT_PATH", "client.pfx")
	t.Setenv("CLIENT_CERT_PASSWORD", "")

	if _, err := LoadConfiguredClient(); err == nil {
		t.Fatal("LoadConfiguredClient() returned nil error, want error")
	}
}

func TestApplyAuthUsesPasswordHeader(t *testing.T) {
	t.Helper()

	client, err := NewClient(&Config{
		BaseURL:  "https://example.cybozu.com/g",
		Username: "user",
		Password: "pass",
	})
	if err != nil {
		t.Fatalf("NewClient() returned error: %v", err)
	}

	req, err := http.NewRequest("GET", client.GetBaseURL(), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}

	client.ApplyAuth(req)
	if got := req.Header.Get("X-Cybozu-Authorization"); got != "dXNlcjpwYXNz" {
		t.Fatalf("X-Cybozu-Authorization = %q, want %q", got, "dXNlcjpwYXNz")
	}
}

func TestApplyAuthUsesBearerHeader(t *testing.T) {
	t.Helper()

	client, err := NewClient(&Config{
		BaseURL:     "https://example.cybozu.com/g",
		AuthType:    AuthTypeOAuth,
		BearerToken: "token",
	})
	if err != nil {
		t.Fatalf("NewClient() returned error: %v", err)
	}

	req, err := http.NewRequest("GET", client.GetBaseURL(), nil)
	if err != nil {
		t.Fatalf("http.NewRequest() returned error: %v", err)
	}

	client.ApplyAuth(req)
	if got := req.Header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer token")
	}
}
