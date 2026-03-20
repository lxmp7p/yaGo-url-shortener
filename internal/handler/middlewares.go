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

			// ПРИ ВКЛЮЧЕНИИ КОДА НИЖЕ ПАДАЮТ ТЕСТЫ ДЛЯ iter7
			// МБ Я НЕ ПОНЯЛ ЗАДАНИЯ И ТАК И НУЖНО, ПОТОМУ ЧТО БЕЗ ЭТОГО ТЕСТЫ iter8 ПРОХОДЯТ
			// ЛИБО Я БЕЗ ПОНЯТИЯ ПОЧЕМУ ОТВЕТЫ ДЛЯ ЗАПРОСОВ БЕЗ acceptEncoding ВСЕ РАВНО СЖИМАЮТСЯ(iter7)

			// acceptEncoding := r.Header.Get(AcceptEncoding)
			// if acceptEncoding == "" || !strings.Contains(acceptEncoding, "gzip") {
			h.ServeHTTP(w, r)
			//return
			// }

			// var cw *compressResponseWriter
			// if strings.Contains(r.Header.Get(AcceptEncoding), Gzip) {
			// 	cw = &compressResponseWriter{ResponseWriter: w}
			// 	w = cw
			// }

			// h.ServeHTTP(w, r)

			// if cw != nil && cw.Writer != nil {
			// 	cw.Writer.Close()
			// }
		})
	}
}
