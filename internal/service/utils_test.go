package service

import (
	"context"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	key := UserIDContextKey()
	ctx := context.WithValue(context.Background(), key, "ttt")
	id, ok := UserIDFromContext(ctx)
	if !ok {
		t.Fatal("expected ok = true")
	}
	if id != "ttt" {
		t.Fatalf("expected test-user, got %s", id)
	}
}

func TestGenerateUserCookie(t *testing.T) {
	cookie := GenerateUserCookie("ttt1", "ttt2")
	if cookie.Value != "ttt1:ttt2" {
		t.Fatalf("wrong value: %s", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Fatal("HttpOnly must be true")
	}
	if cookie.Path != "/" {
		t.Fatal("wrong path")
	}
}

func TestGenerateShortURL(t *testing.T) {
	got := generateShortURL()
	if len(got) != 8 {
		t.Fatalf("expected length 8, got %d", len(got))
	}
}
