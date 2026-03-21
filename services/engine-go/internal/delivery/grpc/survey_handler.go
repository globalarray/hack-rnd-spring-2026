package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sourcecraft.dev/benzo/testengine/internal/gen/pb"
	"sourcecraft.dev/benzo/testengine/internal/service/survey/dto"
)

type surveyHandler struct {
	pb.UnimplementedSurveyAdminServiceServer

	log     *slog.Logger
	service surveyService
}

type surveyService interface {
	Create(ctx context.Context, input *dto.CreateSurveyInput) (string, error)
}

func RegisterSurveyAdminServiceServer(grpcServer *grpc.Server, log *slog.Logger, service surveyService) {
	pb.RegisterSurveyAdminServiceServer(grpcServer, &surveyHandler{log: log, service: service})
}

func (sh *surveyHandler) CreateSurvey(ctx context.Context, req *pb.CreateSurveyRequest) (*pb.CreateSurveyResponse, error) {
	const op = "surveyHandler.CreateSurvey"

	if err := req.Validate(); err != nil {
		sh.log.With(slog.String("op", op), slog.Any("error", err)).Error("validate error")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	input := mapCreateSurveyRequestToInput(req)

	uuid, err := sh.service.Create(ctx, input)

	if err != nil {
		sh.log.With(slog.String("op", op), slog.Any("error", err)).Error("create uuid")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateSurveyResponse{
		SurveyId: uuid,
	}, nil
}
