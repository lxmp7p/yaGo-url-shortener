package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr            string
	ResultAddr      string
	FileStoragePath string
	DatabaseDsn     string
	FileLogging     LoggerInfo
	RemoteLogging   LoggerInfo
}

type LoggerInfo struct {
	Path   string
	Enable bool
}

func (li *LoggerInfo) setPath(path string) {
	li.Enable = true
	li.Path = path
}

func NewConfig() Config {
	return Config{}
}

func (cfg *Config) InitConfig() Config {
	cfg.argsConfigurator()
	cfg.envConfigurator()

	return Config{
		Addr:            cfg.Addr,
		ResultAddr:      cfg.ResultAddr,
		FileStoragePath: cfg.FileStoragePath,
		DatabaseDsn:     cfg.DatabaseDsn,
		RemoteLogging:   cfg.RemoteLogging,
		FileLogging:     cfg.FileLogging,
	}
}

func (cfg *Config) argsConfigurator() {
	addr := flag.String("a", "localhost:8080", "server ip:port")
	resultAddr := flag.String("b", "http://localhost:8080", "server result ip:port")
	fileStoragePath := flag.String("f", "storageFile", "file storage path")
	databaseDsn := flag.String("d", "", "database connection string")
	fileLoggingPath := flag.String("audit-file", "", "file logging path")
	remoteLoggingURL := flag.String("audit-url", "", "remote logging url")

	flag.Parse()

	cfg.Addr = *addr
	cfg.ResultAddr = *resultAddr
	cfg.FileStoragePath = *fileStoragePath
	cfg.DatabaseDsn = *databaseDsn

	if *fileLoggingPath != "" {
		cfg.FileLogging.setPath(*fileLoggingPath)
	}

	if *remoteLoggingURL != "" {
		cfg.RemoteLogging.setPath(*remoteLoggingURL)
	}
}

func (cfg *Config) envConfigurator() {
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envResultAddr := os.Getenv("BASE_URL"); envResultAddr != "" {
		cfg.ResultAddr = envResultAddr
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envDatabaseDsn := os.Getenv("DATABASE_DSN"); envDatabaseDsn != "" {
		cfg.DatabaseDsn = envDatabaseDsn
	}
	if envFileLoggingPath := os.Getenv("AUDIT_FILE"); envFileLoggingPath != "" {
		cfg.FileLogging.setPath(envFileLoggingPath)
	}
	if envRemoteLoggingPath := os.Getenv("AUDIT_URL"); envRemoteLoggingPath != "" {
		cfg.RemoteLogging.setPath(envRemoteLoggingPath)
	}
}
