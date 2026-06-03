package service

import "testing"

func TestShortenerService_SignVerify_OK(t *testing.T) {
	s := &ShortenerService{
		secret: []byte("secret"),
	}
	userID := "user123"

	signature := s.Sign(userID)
	if signature == "" {
		t.Fatal("signature should not be empty")
	}
	if !s.Verify(userID, signature) {
		t.Fatal("expected signature to be valid")
	}
}
