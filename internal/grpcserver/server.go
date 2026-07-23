package grpcserver

import (
	"context"
	"errors"

	"github.com/lxmp7p/yaGo-url-shortener/internal/repository"
	"github.com/lxmp7p/yaGo-url-shortener/internal/service"
	"github.com/lxmp7p/yaGo-url-shortener/proto/shortener"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server - gRPC-фасад над бизнес-логикой ShortenerService.
type Server struct {
	shortener.UnimplementedShortenerServiceServer
	Service *service.ShortenerService
}

func NewServer(s *service.ShortenerService) *Server {
	return &Server{Service: s}
}

func (s *Server) ShortenURL(ctx context.Context, req *shortener.URLShortenRequest) (*shortener.URLShortenResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url is empty")
	}

	result, _, err := s.Service.ShortenURL(ctx, req.GetUrl(), userID)
	if err != nil {
		s.Service.Logger.Error(err.Error())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return shortener.URLShortenResponse_builder{
		Result: result,
	}.Build(), nil
}

func (s *Server) ExpandURL(ctx context.Context, req *shortener.URLExpandRequest) (*shortener.URLExpandResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is empty")
	}

	original, err := s.Service.ExpandURL(ctx, req.GetId(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrURLDeleted) {
			return nil, status.Error(codes.NotFound, "url deleted")
		}
		if errors.Is(err, repository.ErrURLNotFound) {
			return nil, status.Error(codes.NotFound, "url not found")
		}
		s.Service.Logger.Error(err.Error())
		return nil, status.Error(codes.Internal, "internal error")
	}

	return shortener.URLExpandResponse_builder{
		Result: original,
	}.Build(), nil
}

func (s *Server) ListUserURLs(ctx context.Context, _ *shortener.ListUserURLsRequest) (*shortener.UserURLsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	urls, err := s.Service.ListUserURLs(ctx, userID)
	if err != nil {
		s.Service.Logger.Error(err.Error())
		return nil, status.Error(codes.Internal, "internal error")
	}

	var items []*shortener.URLData

	for _, u := range urls {
		items = append(items, shortener.URLData_builder{
			ShortUrl:    u.Short,
			OriginalUrl: u.Original,
		}.Build())
	}

	return shortener.UserURLsResponse_builder{
		Url: items,
	}.Build(), nil
}
