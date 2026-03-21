package analytics

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sourcecraft.dev/benzo/testengine/internal/domain"
	"sourcecraft.dev/benzo/testengine/internal/gen/pb"
	servicedto "sourcecraft.dev/benzo/testengine/internal/service/analytics/dto"
)

type analyticsHandler struct {
	pb.UnimplementedAnalyticsServiceServer
	log     *slog.Logger
	service analyticsService
}

type analyticsService interface {
	GetSessionData(ctx context.Context, input *servicedto.GetSessionDataInput) (*servicedto.GetSessionDataOutput, error)
}

func RegisterAnalyticsServiceServer(server *grpc.Server, log *slog.Logger, service analyticsService) {
	pb.RegisterAnalyticsServiceServer(server, &analyticsHandler{log: log, service: service})
}

func (h *analyticsHandler) GetSessionDataForAnalytics(ctx context.Context, req *pb.GetSessionDataRequest) (*pb.SessionAnalyticsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	out, err := h.service.GetSessionData(ctx, &servicedto.GetSessionDataInput{SessionID: req.SessionId})
	if err != nil {
		h.log.With(slog.Any("error", err)).Error("get session data for analytics")
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return mapGetSessionDataOutput(out), nil
}
