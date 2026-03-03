package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	"github.com/lxmp7p/yaGo-url-shortener/internal/config"
	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShortURLHandler(t *testing.T) {
	var tmpURL string
	var templateURL = "http://localhost:8080/"

	app := App{
		Config: config.Config{
			Addr:       "localhost:8080",
			ResultAddr: "http://localhost:8080",
		},
		Storage: repository.NewCache(),
	}

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		method      string
		path        string
		body        *bytes.Reader
		contentType string
		want        want
	}{
		{
			name:        "negative test #1 POST",
			method:      http.MethodPost,
			path:        "/",
			body:        bytes.NewReader([]byte("yandex.ru")),
			contentType: "biba",
			want: want{
				code:        415,
				response:    "Unsupported Media Type\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "positive test #1 POST",
			method:      http.MethodPost,
			path:        "/",
			body:        bytes.NewReader([]byte("yandex.ru")),
			contentType: "text/plain",
			want: want{
				code:        201,
				response:    templateURL,
				contentType: "text/plain",
			},
		},
		{
			name:        "positive test #1 GET",
			method:      http.MethodGet,
			path:        "/",
			body:        bytes.NewReader([]byte("yandex.ru")),
			contentType: "biba",
			want: want{
				code:        307,
				response:    "",
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.method == http.MethodGet {
				test.path = fmt.Sprintf("/%s", tmpURL)
			}
			request := httptest.NewRequest(
				test.method,
				test.path,
				test.body,
			)
			request.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()

			router := InitRoutes(app)

			router.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Contains(t, string(resBody), test.want.response)
			if test.method == http.MethodPost &&
				test.want.code == http.StatusCreated {
				if assert.Contains(t, string(resBody), templateURL) {
					u, err := url.Parse(string(resBody))
					require.NoError(t, err)
					tmpURL = path.Base(u.Path)
				}
			}
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
