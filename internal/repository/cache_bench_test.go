package repository

import (
	"context"
	"os"
	"testing"
)

func BenchmarkCacheSaveTest(b *testing.B) {
	tmpFile, _ := os.CreateTemp("", "cache")
	defer os.Remove(tmpFile.Name())

	cache := &Cache{
		URLCache: make(map[string]URL, 1000),
		file:     tmpFile,
	}

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = cache.Save(
			ctx,
			"minecraft.com",
			"short",
			"user",
		)
	}
}

func BenchmarkCacheLoadTest(b *testing.B) {
	tmpFile, _ := os.CreateTemp("", "cache")
	defer os.Remove(tmpFile.Name())

	cache := &Cache{
		URLCache: make(map[string]URL, 1000),
		file:     tmpFile,
	}

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = cache.Load(
			ctx,
			"short",
		)
	}
}

func BenchmarkCacheGetTest(b *testing.B) {
	tmpFile, _ := os.CreateTemp("", "cache")
	defer os.Remove(tmpFile.Name())

	cache := &Cache{
		URLCache: make(map[string]URL, 1000),
		file:     tmpFile,
	}

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = cache.Get(
			ctx,
			"short",
		)
	}
}
