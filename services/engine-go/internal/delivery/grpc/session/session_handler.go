package session

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sourcecraft.dev/benzo/testengine/internal/domain"
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
	Start(ctx context.Context, input *dto.StartSessionInput) (*dto.StartSessionOutput, error)
	SubmitAnswer(context.Context, dto.SubmitAnswerInput) (*dto.SubmitAnswerOutput, error)
	CurrentQuestion(ctx context.Context, sessionID string) (*question.Question, error)
}

func RegisterSessionClientServer(server *grpc.Server, log *slog.Logger, service sessionService) {
	pb.RegisterSessionClientServiceServer(server, &sessionHandler{log: log, service: service})
}

func (s *sessionHandler) StartSession(ctx context.Context, req *pb.StartSessionRequest) (*pb.StartSessionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	q, err := s.service.Start(ctx, mapStartSessionRequestToStartSessionInput(req))

	if err != nil {
		if errors.Is(err, domain.ErrShareLinkUsed) {
			return nil, status.Error(codes.AlreadyExists, "Эта ссылка уже использована. Создайте новую ссылку для следующего прохождения.")
		}
		if errors.Is(err, domain.ErrConflict) {
			return nil, status.Error(codes.AlreadyExists, "session already started")
		}
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return mapDomainToStartSessionResponse(q), nil
}

func (s *sessionHandler) SubmitAnswer(ctx context.Context, req *pb.SubmitAnswerRequest) (*pb.SubmitAnswerResponse, error) {
	if req.GetSessionId() == "" || req.GetQuestionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id and question_id are required")
	}

	input := dto.SubmitAnswerInput{
		SessionID:  req.GetSessionId(),
		QuestionID: req.GetQuestionId(),
	}

	switch payload := req.Payload.(type) {
	case *pb.SubmitAnswerRequest_AnswerId:
		input.AnswerID = payload.AnswerId
	case *pb.SubmitAnswerRequest_RawText:
		input.RawText = payload.RawText
	case *pb.SubmitAnswerRequest_MultipleChoice:
		return nil, status.Error(codes.Unimplemented, "multiple_choice payload is not supported yet")
	default:
		return nil, status.Error(codes.InvalidArgument, "answer payload is missing")
	}

	output, err := s.service.SubmitAnswer(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConflict):
			return nil, status.Error(codes.AlreadyExists, "answer already submitted or session state changed")
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "question, answer, or session not found")
		case errors.Is(err, domain.ErrTimeLimitExceeded):
			return nil, status.Error(codes.FailedPrecondition, "session time limit exceeded")
		default:
			return nil, status.Error(codes.Internal, "failed to process answer")
		}
	}

	return mapSubmitAnswerResponse(output), nil
}

func (s *sessionHandler) GetCurrentQuestion(ctx context.Context, req *pb.GetSessionCurrentQuestionRequest) (*pb.QuestionClientView, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	q, err := s.service.CurrentQuestion(ctx, req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "active question not found for this session")
	}

	return mapQuestionToQuestionClientView(q), err
}
