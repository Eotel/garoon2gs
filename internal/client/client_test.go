package client

import "testing"

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
