package main

import "testing"

func TestValidateAppEnvSuccess(t *testing.T) {
	t.Helper()

	t.Setenv("SPREADSHEET_ID", "spreadsheet-id")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "service-account.json")
	t.Setenv("USER_MAPPING_PATH", "user_mapping.csv")

	if err := validateAppEnv(); err != nil {
		t.Fatalf("validateAppEnv() returned error: %v", err)
	}
}

func TestValidateAppEnvMissingValues(t *testing.T) {
	t.Helper()

	t.Setenv("SPREADSHEET_ID", "")
	t.Setenv("GOOGLE_SERVICE_ACCOUNT_FILE", "")
	t.Setenv("USER_MAPPING_PATH", "")

	if err := validateAppEnv(); err == nil {
		t.Fatal("validateAppEnv() returned nil, want error")
	}
}
