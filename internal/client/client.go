package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/pkcs12"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// GaroonClient represents a client for Garoon API
type GaroonClient struct {
	BaseURL      string
	Username     string
	Password     string
	CertPath     string
	CertPassword string
	Client       *http.Client
}

// Config holds client configuration
type Config struct {
	BaseURL      string
	Username     string
	Password     string
	CertPath     string
	CertPassword string
}

// getConfigDir returns the configuration directory
func getConfigDir() (string, error) {
	// まず実行ファイルのディレクトリを試す
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			return dir, nil
		}
	}

	// 次にカレントディレクトリを試す
	if pwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(pwd, ".env")); err == nil {
			return pwd, nil
		}
	}

	// 最後にカレントディレクトリを返す（.envが見つからなくても）
	return os.Getwd()
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("設定ディレクトリの取得に失敗しました: %v", err)
	}

	if err := godotenv.Load(filepath.Join(configDir, ".env")); err != nil {
		log.Println("Warning: .env ファイルが見つかりませんでした")
	}

	baseURL := os.Getenv("GAROON_BASE_URL")
	username := os.Getenv("GAROON_USERNAME")
	password := os.Getenv("GAROON_PASSWORD")

	if baseURL == "" || username == "" || password == "" {
		return nil, fmt.Errorf("必要な環境変数が設定されていません")
	}

	return &Config{
		BaseURL:      baseURL,
		Username:     username,
		Password:     password,
		CertPath:     filepath.Join(configDir, os.Getenv("CLIENT_CERT_PATH")),
		CertPassword: os.Getenv("CLIENT_CERT_PASSWORD"),
	}, nil
}

// NewClient creates a new GaroonClient instance
func NewClient(config *Config) (*GaroonClient, error) {
	var httpClient *http.Client

	if config.CertPath != "" && config.CertPassword != "" {
		// クライアント証明書の設定
		pfxData, err := os.ReadFile(config.CertPath)
		if err != nil {
			return nil, fmt.Errorf("証明書の読み込みに失敗しました: %v", err)
		}

		blocks, err := pkcs12.ToPEM(pfxData, config.CertPassword)
		if err != nil {
			return nil, fmt.Errorf("証明書の解析に失敗しました: %v", err)
		}

		var cert tls.Certificate
		for _, b := range blocks {
			if b.Type == "CERTIFICATE" {
				cert.Certificate = append(cert.Certificate, b.Bytes)
			}
			if b.Type == "PRIVATE KEY" {
				cert.PrivateKey, err = x509.ParsePKCS1PrivateKey(b.Bytes)
				if err != nil {
					return nil, fmt.Errorf("秘密鍵の解析に失敗しました: %v", err)
				}
			}
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			},
			Timeout: 10 * time.Second,
		}
	} else {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	return &GaroonClient{
		BaseURL:      config.BaseURL,
		Username:     config.Username,
		Password:     config.Password,
		CertPath:     config.CertPath,
		CertPassword: config.CertPassword,
		Client:       httpClient,
	}, nil
}
