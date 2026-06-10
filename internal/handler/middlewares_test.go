package handler

import (
	"net/http/httptest"
	"testing"
)

func TestCompressResponseWriter_GzipEnabled(t *testing.T) {

	resp := httptest.NewRecorder()

	w := &compressResponseWriter{
		ResponseWriter: resp,
	}

	w.Header().Set(ContentType, "text/html")

	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	if n == 0 {
		t.Fatal("expected bytes written")
	}

	if resp.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip header")
	}
}
