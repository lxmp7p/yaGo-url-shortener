package handler

import (
	"compress/gzip"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"github.com/sirupsen/logrus"
)

// generate:reset
type logRequest struct {
	URI          string
	Method       string
	Time         time.Duration
	StatusCode   int
	ResponseSize int
}

func (lr logRequest) String() string {
	return fmt.Sprintf(
		"method:%s uri:%s code:%d duration:%s size:%d",
		lr.Method, lr.URI, lr.StatusCode, lr.Time, lr.ResponseSize,
	)
}

// generate:reset
type loggerResponseWriter struct {
	http.ResponseWriter
	size       int
	statusCode int
}

func (lw *loggerResponseWriter) Write(b []byte) (int, error) {
	n, err := lw.ResponseWriter.Write(b)
	lw.size += n
	return n, err
}

func (lw *loggerResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

// generate:reset
type compressResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (compressWriter *compressResponseWriter) Write(b []byte) (int, error) {
	contentType := compressWriter.Header().Get(ContentType)
	if compressWriter.Writer == nil {
		compressWriter.Header().Set("Content-Encoding", "gzip")
		if strings.Contains(contentType, AppJSON) || strings.Contains(contentType, TextHTML) {
			compressWriter.Writer = gzip.NewWriter(compressWriter.ResponseWriter)
		}
	}

	if compressWriter.Writer != nil {
		return compressWriter.Writer.Write(b)
	}

	return compressWriter.ResponseWriter.Write(b)
}

func LoggingMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logData := logRequest{
				URI:    r.RequestURI,
				Method: r.Method,
			}
			loggerRespWriter := &loggerResponseWriter{ResponseWriter: w}
			h.ServeHTTP(loggerRespWriter, r)

			logData.ResponseSize = loggerRespWriter.size
			logData.StatusCode = loggerRespWriter.statusCode
			logData.Time = time.Since(start)

			logger.Info(logData)
		})
	}
}

func CompressMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get(ContentEncoding), Gzip) {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				r.Body = gz
				defer gz.Close()
			}

			h.ServeHTTP(w, r)
		})
	}
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user")
		if err != nil {
			userID := uuid.NewString()
			signature := h.Service.Sign(userID)

			http.SetCookie(w, service.GenerateUserCookie(userID, signature))
			userIDCtx := service.UserIDContextKey()
			ctx := context.WithValue(r.Context(), userIDCtx, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		parts := strings.Split(cookie.Value, ":")
		if len(parts) != 2 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userID, signature := parts[0], parts[1]

		if !h.Service.Verify(userID, signature) {
			userID := uuid.NewString()
			signature := h.Service.Sign(userID)
			http.SetCookie(w, service.GenerateUserCookie(userID, signature))
		}
		userIDCtx := service.UserIDContextKey()

		ctx := context.WithValue(r.Context(), userIDCtx, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TrustedSubnetMiddleware проверяет, что IP-адрес из заголовка X-Real-IP
// входит в доверенную подсеть trustedSubnet. Если подсеть не задана -
// доступ запрещён для всех запросов.
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			if err != nil {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(realIP)
			if ip == nil || !ipNet.Contains(ip) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
