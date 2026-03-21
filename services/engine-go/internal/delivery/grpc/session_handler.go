package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
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
