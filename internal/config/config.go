package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr       string
	ResultAddr string
}

func NewConfig() Config {
	return Config{}
}

func (cfg *Config) InitConfig() Config {
	addr := flag.String("a", "localhost:8080", "server ip:port")
	resultAddr := flag.String("b", "http://localhost:8080", "server result ip:port")

	cfg.envConfigurator()

	flag.Parse()

	return Config{
		Addr:       *addr,
		ResultAddr: *resultAddr,
	}
}

func (cfg *Config) envConfigurator() {
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.Addr = envAddr
	}
	if envResultAddr := os.Getenv("BASE_URL"); envResultAddr != "" {
		cfg.ResultAddr = envResultAddr
	}
}
