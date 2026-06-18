package main

import (
	"log"

	"github.com/lxmp7p/yaGo-url-shortener/internal/generator"
)

func main() {
	if err := generator.Run("."); err != nil {
		log.Fatal(err)
	}
}
