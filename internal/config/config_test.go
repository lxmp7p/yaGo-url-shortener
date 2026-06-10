package config

import (
	"testing"
)

func TestConfig_InitConfig_envOverride(t *testing.T) {
	t.Setenv("SERVER_ADDRESS", "127.0.0.1:9000")
	t.Setenv("BASE_URL", "http://test")
	t.Setenv("FILE_STORAGE_PATH", "/tmp/file")
	t.Setenv("DATABASE_DSN", "postgres://test")

	cfg := NewConfig()
	result := cfg.InitConfig()

	if result.Addr != "127.0.0.1:9000" {
		t.Fatalf("unexpected Addr: %s", result.Addr)
	}
	if result.ResultAddr != "http://test" {
		t.Fatalf("unexpected ResultAddr: %s", result.ResultAddr)
	}
	if result.FileStoragePath != "/tmp/file" {
		t.Fatalf("unexpected FileStoragePath: %s", result.FileStoragePath)
	}
	if result.DatabaseDsn != "postgres://test" {
		t.Fatalf("unexpected DatabaseDsn: %s", result.DatabaseDsn)
	}
}
