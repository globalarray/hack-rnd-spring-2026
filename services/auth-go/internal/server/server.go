package server

import (
	"context"
	"database/sql"
	"errors"
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
	authService *service.AuthService
}

func New(authService *service.AuthService) *AuthServer {
	return &AuthServer{authService: authService}
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

func mapDirectoryEntryToResponse(item postgres.DirectoryEntry) *auth.DirectoryEntry {
	response := &auth.DirectoryEntry{
		Id:              item.ID,
		Email:           item.Email,
		FullName:        item.FullName,
		Phone:           item.Phone,
		Role:            item.Role,
		Status:          item.Status,
		InvitationToken: item.InvitationToken,
	}

	if !item.AccessUntil.IsZero() {
		response.AccessUntil = timestamppb.New(item.AccessUntil)
	}
	if !item.ExpiresAt.IsZero() {
		response.ExpiresAt = timestamppb.New(item.ExpiresAt)
	}

	return response
}

func (s *AuthServer) ListPsychologists(ctx context.Context, _ *emptypb.Empty) (*auth.ListPsychologistsResponse, error) {
	if _, err := s.authService.Authorize(ctx, true); err != nil {
		return nil, err
	}

	items, err := s.authService.ListPsychologists(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list psychologists")
	}

	response := &auth.ListPsychologistsResponse{
		Items: make([]*auth.DirectoryEntry, 0, len(items)),
	}
	for _, item := range items {
		response.Items = append(response.Items, mapDirectoryEntryToResponse(item))
	}

	return response, nil
}

func (s *AuthServer) CreateInvitation(ctx context.Context, in *auth.CreateInvitationRequest) (*auth.InvitationResponse, error) {
	if _, err := s.authService.Authorize(ctx, true); err != nil {
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

	token, err := s.authService.CreateInvitation(ctx, email, role, fullName, phone, accessUntil, expiresAt)
	if errors.Is(err, service.ErrUserAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
	}
	if err != nil {
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

	tokens, role, err := s.authService.Register(ctx, token, password)
	if err != nil {
		switch err.Error() {
		case service.ErrAlreadyUsed:
			return nil, status.Error(codes.AlreadyExists, service.ErrAlreadyUsed)
		case service.ErrExpiredToken:
			return nil, status.Error(codes.FailedPrecondition, service.ErrExpiredToken)
		case service.ErrTokenNotExists:
			return nil, status.Error(codes.NotFound, service.ErrTokenNotExists)
		default:
			if errors.Is(err, sql.ErrTxDone) {
				return nil, status.Error(codes.Internal, "failed to commit transaction")
			}
			return nil, status.Error(codes.Internal, "failed to register user")
		}
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

	tokens, role, err := s.authService.Login(ctx, email, password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		if errors.Is(err, service.ErrAccountBlocked) {
			return nil, status.Error(codes.PermissionDenied, "account is blocked")
		}
		if errors.Is(err, service.ErrAccountInactive) {
			return nil, status.Error(codes.PermissionDenied, "account is inactive")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &auth.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.Expires_in,
		Role:         role,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, in *auth.RefreshTokenRequest) (*auth.TokenResponse, error) {
	refreshToken := strings.TrimSpace(in.GetRefreshToken())
	if refreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	newTokens, role, err := s.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidSession) {
			return nil, status.Error(codes.Unauthenticated, "invalid session")
		}
		if errors.Is(err, service.ErrAccountBlocked) {
			return nil, status.Error(codes.PermissionDenied, "account is blocked")
		}
		if errors.Is(err, service.ErrAccountInactive) {
			return nil, status.Error(codes.PermissionDenied, "account is inactive")
		}
		return nil, status.Error(codes.Internal, "failed to generate new tokens")
	}

	return &auth.TokenResponse{
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
		ExpiresIn:    newTokens.Expires_in,
		Role:         role,
	}, nil
}

func (s *AuthServer) BlockUser(ctx context.Context, in *auth.BlockUserRequest) (*emptypb.Empty, error) {
	if _, err := s.authService.Authorize(ctx, true); err != nil {
		return nil, err
	}

	if err := s.authService.BlockUser(ctx, strings.TrimSpace(in.GetEmail())); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to block user")
	}

	return &emptypb.Empty{}, nil
}

func (s *AuthServer) UnBlockUser(ctx context.Context, in *auth.UnBlockUserRequest) (*emptypb.Empty, error) {
	if _, err := s.authService.Authorize(ctx, true); err != nil {
		return nil, err
	}

	if err := s.authService.UnblockUser(ctx, strings.TrimSpace(in.GetEmail())); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to unblock user")
	}

	return &emptypb.Empty{}, nil
}

func (s *AuthServer) GetProfile(ctx context.Context, _ *emptypb.Empty) (*auth.ProfileResponse, error) {
	authState, err := s.authService.Authorize(ctx, false)
	if err != nil {
		return nil, err
	}

	user, err := s.authService.GetProfile(ctx, authState.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return mapUserToProfileResponse(user), nil
}

func (s *AuthServer) UpdateProfile(ctx context.Context, in *auth.UpdateProfileRequest) (*auth.ProfileResponse, error) {
	authState, err := s.authService.Authorize(ctx, false)
	if err != nil {
		return nil, err
	}

	user, err := s.authService.UpdateProfile(ctx, authState.ID, in.GetAbout(), in.GetPhotoUrl())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return mapUserToProfileResponse(user), nil
}

func (s *AuthServer) UpdateUserProfile(ctx context.Context, in *auth.UpdateUserProfileRequest) (*auth.ProfileResponse, error) {
	if _, err := s.authService.Authorize(ctx, true); err != nil {
		return nil, err
	}

	userID := strings.TrimSpace(in.GetUserId())
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.authService.UpdateProfile(ctx, userID, in.GetAbout(), in.GetPhotoUrl())
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

	profile, err := s.authService.GetPublicProfile(ctx, userID)
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
