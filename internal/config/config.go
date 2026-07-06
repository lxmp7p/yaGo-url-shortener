package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
)

// generate:reset
type Config struct {
	Addr            string `json:"server_address"`
	ResultAddr      string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	FileLogging     LoggerInfo
	RemoteLogging   LoggerInfo
}

// generate:reset
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
	cfg.setDefaults()

	configPath := cfg.getConfigPath()

	if configPath != "" {
		if err := cfg.loadConfig(configPath); err != nil {
			log.Fatal(err)
		}
	}

	cfg.envConfigurator()
	cfg.argsConfigurator()

	return *cfg
}

func (cfg *Config) setDefaults() {
	cfg.Addr = "localhost:8080"
	cfg.ResultAddr = "http://localhost:8080"
	cfg.FileStoragePath = "storageFile"
}

func (cfg *Config) getConfigPath() string {
	path := os.Getenv("CONFIG")

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	fs.StringVar(&path, "c", path, "config path")
	fs.StringVar(&path, "config", path, "config path")

	_ = fs.Parse(os.Args[1:])

	return path
}

func (cfg *Config) argsConfigurator() {
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "server ip:port")
	flag.StringVar(&cfg.ResultAddr, "b", cfg.ResultAddr, "base url")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "storage path")
	flag.StringVar(&cfg.DatabaseDsn, "d", cfg.DatabaseDsn, "database dsn")

	fileLogging := flag.String("audit-file", "", "file logging path")
	remoteLogging := flag.String("audit-url", "", "remote logging url")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "enable https")

	flag.Parse()

	if *fileLogging != "" {
		cfg.FileLogging.setPath(*fileLogging)
	}

	if *remoteLogging != "" {
		cfg.RemoteLogging.setPath(*remoteLogging)
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
	value := os.Getenv("ENABLE_HTTPS")
	if value != "" {
		enableHTTPS, err := strconv.ParseBool(value)
		if err != nil {
			log.Fatalf("invalid ENABLE_HTTPS: %v", err)
		}
		cfg.EnableHTTPS = enableHTTPS
	}
}

func (cfg *Config) loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fc Config
	if err := json.Unmarshal(data, &fc); err != nil {
		return err
	}

	if fc.Addr != "" {
		cfg.Addr = fc.Addr
	}
	if fc.ResultAddr != "" {
		cfg.ResultAddr = fc.ResultAddr
	}
	if fc.FileStoragePath != "" {
		cfg.FileStoragePath = fc.FileStoragePath
	}
	if fc.DatabaseDsn != "" {
		cfg.DatabaseDsn = fc.DatabaseDsn
	}
	if fc.EnableHTTPS {
		cfg.EnableHTTPS = fc.EnableHTTPS
	}

	return nil
}
