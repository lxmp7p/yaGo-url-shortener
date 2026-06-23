package a

import (
	"log"
	"os"
)

func f() {
	panic("boom")    // want "forbiden panic"
	log.Fatal("err") // want "log.Fatal is forbidden outside main"
	os.Exit(1)       // want "os exit"
}
