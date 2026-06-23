package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/lxmp7p/yaGo-url-shortener/internal/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
