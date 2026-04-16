package repository

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want *Cache
	}{
		{
			name: "success new cache",
			want: &Cache{
				URLCache: make(map[string]URL),
				filename: "tmp",
			},
		},
		{
			name: "success new cache",
			want: &Cache{
				URLCache: make(map[string]URL),
				filename: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "cache")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())
			got := NewCache(context.Background(), tmpFile.Name())
			assert.Equal(t, tmpFile.Name(), got.filename)
			assert.Empty(t, got.URLCache)
		})
	}
}

func TestCache_Save(t *testing.T) {
	tests := []struct {
		name        string // description of this test case
		filepath    string
		originalURL string
		shortURL    string
		wantErr     bool
	}{
		{
			name:        "success save",
			filepath:    "tmp",
			originalURL: "ya.ru",
			shortURL:    "---",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "tmp")
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			cache := NewCache(context.Background(), tmpFile.Name())
			gotErr := cache.Save(context.Background(), tt.originalURL, tt.shortURL, "userId")
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Save() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Save() succeeded unexpectedly")
			}
		})
	}
}

func TestCache_Get(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		filepath string
		shortURL string
		want     string
		wantErr  bool
	}{
		{
			name:     "url not found",
			filepath: "tmp",
			shortURL: "--------",
			want:     "--------",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "tmp")
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			cache := NewCache(context.Background(), tt.filepath)
			got, gotErr := cache.Get(context.Background(), tt.shortURL)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Get() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Get() succeeded unexpectedly")
			}
			if true {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
