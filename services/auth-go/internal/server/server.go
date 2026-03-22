package server

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	auth "sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/gen/go"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/service"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/storage/postgres"
)

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	repo *postgres.Storage
}

func New(repo *postgres.Storage) *AuthServer {
	return &AuthServer{repo: repo}
}

func (s *AuthServer) authorize(ctx context.Context, requireAdmin bool) (*postgres.AuthState, error) {
	token, err := service.ExtractToken(ctx)
	if err != nil {
		return nil, err
	}

	claims, err := service.ValidateAccessToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid token payload")
	}

	authState, err := s.repo.GetUserAuthStateByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	switch authState.Status {
	case postgres.StatusBlocked:
		return nil, status.Error(codes.PermissionDenied, "account is blocked")
	case postgres.StatusInactive:
		return nil, status.Error(codes.PermissionDenied, "account is inactive")
	}

	if requireAdmin && authState.Role != postgres.RoleAdmin {
		return nil, status.Error(codes.PermissionDenied, "admin role is required")
	}

	return authState, nil
}

func mapUserToProfileResponse(user *postgres.User) *auth.ProfileResponse {
	if user == nil {
		return &auth.ProfileResponse{}
	}

	return &auth.ProfileResponse{
		Id:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		Phone:       user.Phone,
		Role:        user.Role,
		PhotoUrl:    user.PhotoURL,
		About:       user.About,
		Status:      user.Status,
		AccessUntil: timestamppb.New(user.AccessUntil),
	}
}

func (s *AuthServer) CreateInvitation(ctx context.Context, in *auth.CreateInvitationRequest) (*auth.InvitationResponse, error) {
	if _, err := s.authorize(ctx, true); err != nil {
		return nil, err
	}

	email := strings.TrimSpace(in.GetEmail())
	fullName := strings.TrimSpace(in.GetFullName())
	phone := strings.TrimSpace(in.GetNumber())
	role := strings.TrimSpace(in.GetRole())
	accessUntilTS := in.GetAccessUntil()
	expiresAtTS := in.GetExpiresAt()

	if email == "" || fullName == "" || phone == "" {
		return nil, status.Error(codes.InvalidArgument, "email, full_name and number are required")
	}
	if accessUntilTS == nil || expiresAtTS == nil {
		return nil, status.Error(codes.InvalidArgument, "access_until and expires_at are required")
	}
	accessUntil := accessUntilTS.AsTime()
	expiresAt := expiresAtTS.AsTime()
	if !expiresAt.After(time.Now()) {
		return nil, status.Error(codes.InvalidArgument, "expires_at must be in the future")
	}

	exists, err := s.repo.UserExistsByEmail(ctx, email)
	if err != nil {
		s.repo.Logger.Error("check invitation user existence failed", slog.Any("error", err), slog.String("email", email))
		return nil, status.Error(codes.Internal, "failed to create invitation")
	}
	if exists {
		return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
	}

	token, err := s.repo.CreateInvitationUUID(ctx, email, role, fullName, phone, accessUntil, expiresAt)
	if err != nil {
		s.repo.Logger.Error("create invitation failed", slog.Any("error", err), slog.String("email", email))
		return nil, status.Error(codes.Internal, "failed to create invitation")
	}

	return &auth.InvitationResponse{Token: token}, nil
}

func (s *AuthServer) Register(ctx context.Context, in *auth.RegisterRequest) (*auth.TokenResponse, error) {
	token := strings.TrimSpace(in.GetToken())
	password := in.GetPassword()

	if token == "" || strings.TrimSpace(password) == "" {
		return nil, status.Error(codes.InvalidArgument, "token and password are required")
	}

	ok, err := service.IsAccessToken(ctx, s.repo, token)
	if !ok {
		switch err.Error() {
		case service.ErrAlreadyUsed:
			return nil, status.Error(codes.AlreadyExists, service.ErrAlreadyUsed)
		case service.ErrExpiredToken:
			return nil, status.Error(codes.FailedPrecondition, service.ErrExpiredToken)
		case service.ErrTokenNotExists:
			return nil, status.Error(codes.NotFound, service.ErrTokenNotExists)
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid invitation token")
		}
	}

	passwordHash, err := service.HashPass(password)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to start transaction")
	}
	defer tx.Rollback()

	id, role, err := s.repo.RegisterByInvite(ctx, tx, token, passwordHash)
	if err != nil {
		s.repo.Logger.Error("register by invite failed", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	tokens, err := service.GenerateTokensTx(ctx, tx, id, role, s.repo)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	if err := tx.Commit(); err != nil {
		return nil, status.Error(codes.Internal, "failed to commit transaction")
	}

	return &auth.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.Expires_in,
		Role:         role,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, in *auth.LoginRequest) (*auth.TokenResponse, error) {
	email := strings.TrimSpace(in.GetEmail())
	password := in.GetPassword()

	if email == "" || strings.TrimSpace(password) == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	switch user.Status {
	case postgres.StatusBlocked:
		return nil, status.Error(codes.PermissionDenied, "account is blocked")
	case postgres.StatusInactive:
		return nil, status.Error(codes.PermissionDenied, "account is inactive")
	}

	if ok := service.CompareHashPass(password, user.PasswordHash); !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid email or password")
	}

	tokens, err := service.GenerateTokens(ctx, user.ID, user.Role, s.repo)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &auth.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.Expires_in,
		Role:         user.Role,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, in *auth.RefreshTokenRequest) (*auth.TokenResponse, error) {
	refreshToken := strings.TrimSpace(in.GetRefreshToken())
	if refreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	claims, err := service.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "refresh token expired or invalid")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid token payload")
	}

	user, err := s.repo.GetUserByRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "invalid session")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	switch user.Status {
	case postgres.StatusBlocked:
		return nil, status.Error(codes.PermissionDenied, "account is blocked")
	case postgres.StatusInactive:
		return nil, status.Error(codes.PermissionDenied, "account is inactive")
	}

	newTokens, err := service.GenerateTokens(ctx, userID, user.Role, s.repo)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate new tokens")
	}

	return &auth.TokenResponse{
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
		ExpiresIn:    newTokens.Expires_in,
		Role:         user.Role,
	}, nil
}

func (s *AuthServer) BlockUser(ctx context.Context, in *auth.BlockUserRequest) (*emptypb.Empty, error) {
	if _, err := s.authorize(ctx, true); err != nil {
		return nil, err
	}

	if err := s.repo.BlockUserByEmail(ctx, strings.TrimSpace(in.GetEmail())); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to block user")
	}

	return &emptypb.Empty{}, nil
}

func (s *AuthServer) UnBlockUser(ctx context.Context, in *auth.UnBlockUserRequest) (*emptypb.Empty, error) {
	if _, err := s.authorize(ctx, true); err != nil {
		return nil, err
	}

	if err := s.repo.UnBlockUserByEmail(ctx, strings.TrimSpace(in.GetEmail())); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to unblock user")
	}

	return &emptypb.Empty{}, nil
}

func (s *AuthServer) GetProfile(ctx context.Context, _ *emptypb.Empty) (*auth.ProfileResponse, error) {
	authState, err := s.authorize(ctx, false)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetProfileByID(ctx, authState.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return mapUserToProfileResponse(user), nil
}

func (s *AuthServer) UpdateProfile(ctx context.Context, in *auth.UpdateProfileRequest) (*auth.ProfileResponse, error) {
	authState, err := s.authorize(ctx, false)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateProfileByID(ctx, authState.ID, in.GetAbout(), in.GetPhotoUrl()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to update profile")
	}

	user, err := s.repo.GetProfileByID(ctx, authState.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return mapUserToProfileResponse(user), nil
}

func (s *AuthServer) UpdateUserProfile(ctx context.Context, in *auth.UpdateUserProfileRequest) (*auth.ProfileResponse, error) {
	if _, err := s.authorize(ctx, true); err != nil {
		return nil, err
	}

	userID := strings.TrimSpace(in.GetUserId())
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if err := s.repo.UpdateProfileByID(ctx, userID, in.GetAbout(), in.GetPhotoUrl()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to update profile")
	}

	user, err := s.repo.GetProfileByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return mapUserToProfileResponse(user), nil
}

func (s *AuthServer) GetPublicProfile(ctx context.Context, in *auth.PublicProfileRequest) (*auth.PublicProfileResponse, error) {
	userID := strings.TrimSpace(in.GetUserId())
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	profile, err := s.repo.GetPublicProfileByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "public profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to get public profile")
	}

	return &auth.PublicProfileResponse{
		FullName: profile.FullName,
		PhotoUrl: profile.PhotoURL,
		About:    profile.About,
	}, nil
}
