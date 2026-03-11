package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr       string
	ResultAddr string
}

func InitConfig() Config {
	addr := flag.String("a", "localhost:8080", "server ip:port")
	resultAddr := flag.String("b", "http://localhost:8080", "server result ip:port")

	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		addr = &envAddr
	}

	if envResultAddr := os.Getenv("BASE_URL"); envResultAddr != "" {
		addr = &envResultAddr
	}

	flag.Parse()

	return Config{
		Addr:       *addr,
		ResultAddr: *resultAddr,
	}
}
