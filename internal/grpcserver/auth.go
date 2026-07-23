package grpcserver

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// AuthInterceptor хранит ссылку на сервис, чтобы проверять/подписывать
// userID так же, как это делает http.AuthMiddleware.
type AuthInterceptor struct {
	Service *service.ShortenerService
}

func NewAuthInterceptor(s *service.ShortenerService) *AuthInterceptor {
	return &AuthInterceptor{Service: s}
}

// если токен в metadata
// отсутствует или невалиден, генерируется новый userID и подпись,
// которые отправляются клиенту через исходящий header "authorization".
func (a *AuthInterceptor) Unary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var userID string

	token := extractToken(ctx)

	if token == "" {
		userID = uuid.NewString()
		a.issueNewToken(ctx, userID)
	} else {
		parts := strings.Split(token, ":")
		if len(parts) != 2 {
			userID = uuid.NewString()
			a.issueNewToken(ctx, userID)
		} else {
			candidateID, signature := parts[0], parts[1]
			if !a.Service.Verify(candidateID, signature) {
				userID = uuid.NewString()
				a.issueNewToken(ctx, userID)
			} else {
				userID = candidateID
			}
		}
	}

	ctx = context.WithValue(ctx, userIDContextKey, userID)
	return handler(ctx, req)
}

func extractToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

// issueNewToken отправляет клиенту новый токен через исходящий metadata-header,
// аналогично http.SetCookie в AuthMiddleware.
func (a *AuthInterceptor) issueNewToken(ctx context.Context, userID string) {
	signature := a.Service.Sign(userID)
	token := userID + ":" + signature
	_ = grpc.SetHeader(ctx, metadata.Pairs("authorization", token))
}

// UserIDFromContext извлекает userID, помещённый интерцептором в контекст.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}
