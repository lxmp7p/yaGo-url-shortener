package config

import (
	"flag"
)

type Config struct {
	Addr       string
	ResultAddr string
}

func InitConfig() Config {
	addr := flag.String("a", "localhost:8080", "server ip:port")
	resultAddr := flag.String("b", "http://localhost:8080", "server result ip:port")

	flag.Parse()

	return Config{
		Addr:       *addr,
		ResultAddr: *resultAddr,
	}
}
