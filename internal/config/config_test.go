package config

import (
	"os"
	"testing"
)

func TestConfig_InitConfig_envOverride(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", "127.0.0.1:9000")
	os.Setenv("BASE_URL", "http://test")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/file")
	os.Setenv("DATABASE_DSN", "postgres://test")

	defer os.Clearenv()
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
