package client

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()

	original, hadOriginal := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%s) returned error: %v", key, err)
	}
	t.Cleanup(func() {
		var err error
		if hadOriginal {
			err = os.Setenv(key, original)
		} else {
			err = os.Unsetenv(key)
		}
		if err != nil {
			t.Fatalf("cleanup for %s returned error: %v", key, err)
		}
	})
}

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
		BaseURL:         "https://example.cybozu.com/g",
		Username:        "user",
		Password:        "pass",
		CertPath:        "/tmp/client.pfx",
		CertPathSet:     true,
		CertPassword:    "",
		CertPasswordSet: false,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() returned nil, want error")
	}
}

func TestConfigValidateAllowsOptionalClientCertificate(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:         "https://example.cybozu.com/g",
		Username:        "user",
		Password:        "pass",
		CertPath:        "/tmp/client.pfx",
		CertPathSet:     true,
		CertPassword:    "secret",
		CertPasswordSet: true,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
}

func TestConfigValidateAllowsEmptyClientCertPassword(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:         "https://example.cybozu.com/g",
		Username:        "user",
		Password:        "pass",
		CertPath:        "/tmp/client.pfx",
		CertPathSet:     true,
		CertPassword:    "",
		CertPasswordSet: true,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
}

func TestConfigValidateRejectsEmptyClientCertPath(t *testing.T) {
	t.Helper()

	cfg := &Config{
		BaseURL:         "https://example.cybozu.com/g",
		Username:        "user",
		Password:        "pass",
		CertPath:        "",
		CertPathSet:     true,
		CertPassword:    "",
		CertPasswordSet: true,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() returned nil, want error")
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
	unsetEnvForTest(t, "CLIENT_CERT_PATH")
	unsetEnvForTest(t, "CLIENT_CERT_PASSWORD")

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
	unsetEnvForTest(t, "CLIENT_CERT_PATH")
	unsetEnvForTest(t, "CLIENT_CERT_PASSWORD")

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
	unsetEnvForTest(t, "CLIENT_CERT_PASSWORD")

	if _, err := LoadConfiguredClient(); err == nil {
		t.Fatal("LoadConfiguredClient() returned nil error, want error")
	}
}

func TestLoadConfigPreservesEmptyClientCertPassword(t *testing.T) {
	t.Helper()

	t.Setenv("CLIENT_CERT_PATH", "client.pfx")
	t.Setenv("CLIENT_CERT_PASSWORD", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if !cfg.CertPathSet {
		t.Fatal("CertPathSet = false, want true")
	}
	if !cfg.CertPasswordSet {
		t.Fatal("CertPasswordSet = false, want true")
	}
	if cfg.CertPassword != "" {
		t.Fatalf("CertPassword = %q, want empty string", cfg.CertPassword)
	}
}

func TestLoadConfigIgnoresAmbientClientCertVarsWhenEnvFileOmitsThem(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	t.Chdir(tempDir)

	envContent := "GAROON_BASE_URL=\"https://example.cybozu.com/g\"\nGAROON_USERNAME=\"user\"\nGAROON_PASSWORD=\"pass\"\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0o644); err != nil {
		t.Fatalf("WriteFile(.env) returned error: %v", err)
	}

	t.Setenv("CLIENT_CERT_PATH", "client.pfx")
	t.Setenv("CLIENT_CERT_PASSWORD", "secret")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if cfg.CertPathSet {
		t.Fatal("CertPathSet = true, want false")
	}
	if cfg.CertPasswordSet {
		t.Fatal("CertPasswordSet = true, want false")
	}
}

func TestLoadConfigReadsEmptyClientCertPasswordFromEnvFile(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	t.Chdir(tempDir)

	envContent := "GAROON_BASE_URL=\"https://example.cybozu.com/g\"\nGAROON_USERNAME=\"user\"\nGAROON_PASSWORD=\"pass\"\nCLIENT_CERT_PATH=\"client.pfx\"\nCLIENT_CERT_PASSWORD=\"\"\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0o644); err != nil {
		t.Fatalf("WriteFile(.env) returned error: %v", err)
	}

	t.Setenv("CLIENT_CERT_PATH", "stale-client.pfx")
	t.Setenv("CLIENT_CERT_PASSWORD", "stale-secret")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned error: %v", err)
	}

	if !cfg.CertPathSet {
		t.Fatal("CertPathSet = false, want true")
	}
	if !cfg.CertPasswordSet {
		t.Fatal("CertPasswordSet = false, want true")
	}
	if cfg.CertPath != filepath.Join(tempDir, "client.pfx") {
		t.Fatalf("CertPath = %q, want %q", cfg.CertPath, filepath.Join(tempDir, "client.pfx"))
	}
	if cfg.CertPassword != "" {
		t.Fatalf("CertPassword = %q, want empty string", cfg.CertPassword)
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
