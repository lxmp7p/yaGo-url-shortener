package config

import (
	"encoding/json"
	"flag"
	"fmt"
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
	TrustedSubnet   string `json:"trusted_subnet"`
	FileLogging     LoggerInfo
	RemoteLogging   LoggerInfo
}

// generate:reset
type LoggerInfo struct {
	Path   string
	Enable bool
}

type cliFlags struct {
	ConfigPath       string
	Addr             string
	ResultAddr       string
	FileStoragePath  string
	DatabaseDsn      string
	EnableHTTPS      bool
	TrustedSubnet    string
	FileLoggingPath  string
	RemoteLoggingURL string
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

	var cli cliFlags

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	cfg.registerFlags(fs, &cli)

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	if cli.ConfigPath != "" {
		if err := cfg.loadConfig(cli.ConfigPath); err != nil {
			log.Fatal(err)
		}
	}

	if err := cfg.envConfigurator(); err != nil {
		log.Fatal(err)
	}

	cfg.applyFlags(fs, &cli)

	return *cfg
}

func (cfg *Config) setDefaults() {
	cfg.Addr = "localhost:8080"
	cfg.ResultAddr = "http://localhost:8080"
	cfg.FileStoragePath = "storageFile"
}

func (cfg *Config) registerFlags(fs *flag.FlagSet, cli *cliFlags) {
	cli.ConfigPath = os.Getenv("CONFIG")

	fs.StringVar(&cli.ConfigPath, "c", cli.ConfigPath, "config path")
	fs.StringVar(&cli.ConfigPath, "config", cli.ConfigPath, "config path")

	fs.StringVar(&cli.Addr, "a", "", "server ip:port")
	fs.StringVar(&cli.ResultAddr, "b", "", "base url")
	fs.StringVar(&cli.FileStoragePath, "f", "", "storage path")
	fs.StringVar(&cli.DatabaseDsn, "d", "", "database dsn")
	fs.BoolVar(&cli.EnableHTTPS, "s", false, "enable https")
	fs.StringVar(&cli.TrustedSubnet, "t", "", "trusted subnet CIDR")

	fs.StringVar(&cli.FileLoggingPath, "audit-file", "", "file logging path")
	fs.StringVar(&cli.RemoteLoggingURL, "audit-url", "", "remote logging url")
}

func (cfg *Config) applyFlags(fs *flag.FlagSet, cli *cliFlags) {
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.Addr = cli.Addr
		case "b":
			cfg.ResultAddr = cli.ResultAddr
		case "f":
			cfg.FileStoragePath = cli.FileStoragePath
		case "d":
			cfg.DatabaseDsn = cli.DatabaseDsn
		case "s":
			cfg.EnableHTTPS = cli.EnableHTTPS
		case "t":
			cfg.TrustedSubnet = cli.TrustedSubnet
		case "audit-file":
			cfg.FileLogging.setPath(cli.FileLoggingPath)
		case "audit-url":
			cfg.RemoteLogging.setPath(cli.RemoteLoggingURL)
		}
	})
}
func (cfg *Config) envConfigurator() error {
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
			return fmt.Errorf("invalid ENABLE_HTTPS: %v", err)
		}
		cfg.EnableHTTPS = enableHTTPS
	}

	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		cfg.TrustedSubnet = envTrustedSubnet
	}

	return nil
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

	cfg.EnableHTTPS = fc.EnableHTTPS

	if fc.TrustedSubnet != "" {
		cfg.TrustedSubnet = fc.TrustedSubnet
	}

	return nil
}
