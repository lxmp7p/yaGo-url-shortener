package handler

import "testing"

func TestHTTPConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ContentEncoding", ContentEncoding, "Content-Encoding"},
		{"ContentType", ContentType, "Content-Type"},
		{"Gzip", Gzip, "gzip"},
		{"AppJSON", AppJSON, "application/json"},
		{"TextHTML", TextHTML, "text/html"},
		{"AcceptEncoding", AcceptEncoding, "Accept-Encoding"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
