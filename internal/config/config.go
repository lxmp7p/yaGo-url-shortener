package config

import (
	"flag"
	"fmt"
)

type Config struct {
	Addr       string
	ResultAddr string
}

func InitConfig() (Config, error) {
	addr := flag.String("a", "localhost:8080", "server ip:port")
	resultAddr := flag.String("b", "http://localhost:8080", "server result ip:port")

	flag.Parse()
	fmt.Println(&addr)
	fmt.Println(&resultAddr)

	return Config{
		Addr:       *addr,
		ResultAddr: *resultAddr,
	}, nil
}
