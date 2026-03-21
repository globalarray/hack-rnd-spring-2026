package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sourcecraft.dev/benzo/testengine/internal/domain/models/question"
	"sourcecraft.dev/benzo/testengine/internal/gen/pb"
	"sourcecraft.dev/benzo/testengine/internal/service/session/dto"
)

type sessionHandler struct {
	pb.UnimplementedSessionClientServiceServer
	log     *slog.Logger
	service sessionService
}

type sessionService interface {
	StartSession(ctx context.Context, input *dto.StartSessionInput) (question.Question, error)
}

func RegisterSessionClientServer(server *grpc.Server, log *slog.Logger, service sessionService) {
	pb.RegisterSessionClientServiceServer(server, &sessionHandler{log: log, service: service})
}

func (s *sessionHandler) StartSession(ctx context.Context, req *pb.StartSessionRequest) (*pb.StartSessionResponse, error) {
	q, err := s.service.StartSession(ctx, mapStartSessionRequestToStartSessionInput(req))

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return mapQuestionToStartSessionResponse(q), nil
}
