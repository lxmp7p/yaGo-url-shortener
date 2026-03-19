package handler

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

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

type compressResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (compressWriter *compressResponseWriter) Write(b []byte) (int, error) {
	return compressWriter.Writer.Write(b)
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

func CompressMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := w
			if strings.Contains(r.Header.Get(ContentEncoding), Gzip) {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				r.Body = gz
				defer gz.Close()
			}

			if checkRequestCompressed(r) {
				compressWriter := gzip.NewWriter(w)
				defer compressWriter.Close()

				writer = &compressResponseWriter{
					ResponseWriter: w,
					Writer:         compressWriter,
				}
			}

			h.ServeHTTP(writer, r)
		})
	}
}

func checkRequestCompressed(r *http.Request) bool {
	return strings.Contains(r.Header.Get(AcceptEncoding), Gzip) &&
		(strings.Contains(r.Header.Get(ContentType), AppJSON) ||
			strings.Contains(r.Header.Get(ContentType), TextHTML))
}
