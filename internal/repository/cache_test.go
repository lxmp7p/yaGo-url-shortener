package repository

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCache(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		filepath string
		want     *Cache
	}{
		{
			name: "success new cache",
			filepath: "tmp",
			want: &Cache{
				URLCache: make(map[string]string),
				filename: "tmp",
			},
		},
		{
			name: "success new cache",
			want: &Cache{
				URLCache: make(map[string]string),
				filename: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCache(tt.filepath)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestCache_Save(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		filepath string
		originalURL string
		shortURL    string
		wantErr     bool
	}{
		{
			name: "success save",
			filepath: "tmp",
			originalURL: "ya.ru",
			shortURL: "---",
			wantErr: false,
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
			
			cache := NewCache(tt.filepath)
			gotErr := cache.Save(tt.originalURL, tt.shortURL)
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
		name string // description of this test case
		filepath string
		shortURL string
		want     string
		wantErr  bool
	}{
		{
			name: "url not found",
			filepath: "tmp",
			shortURL: "--------",
			want: "--------",
			wantErr: true,
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

			cache := NewCache(tt.filepath)
			got, gotErr := cache.Get(tt.shortURL)
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
