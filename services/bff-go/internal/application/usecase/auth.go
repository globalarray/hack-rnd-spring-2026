package usecase

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"sourcecraft.dev/benzo/bff/internal/application/ports"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

type AuthUseCase struct {
	auth          ports.AuthGateway
	publicBaseURL string
}

func NewAuthUseCase(auth ports.AuthGateway, publicBaseURL string) *AuthUseCase {
	return &AuthUseCase{
		auth:          auth,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
	}
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*domain.AuthTokens, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	return uc.auth.Login(ctx, email, password)
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, fmt.Errorf("%w: refreshToken is required", domain.ErrInvalidInput)
	}

	return uc.auth.RefreshToken(ctx, refreshToken)
}

func (uc *AuthUseCase) Register(ctx context.Context, token, password string) (*domain.AuthTokens, error) {
	if _, err := uuid.Parse(strings.TrimSpace(token)); err != nil {
		return nil, fmt.Errorf("%w: token must be a valid UUID", domain.ErrInvalidInput)
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	return uc.auth.Register(ctx, token, password)
}

func (uc *AuthUseCase) GetProfile(ctx context.Context, authorization string) (*domain.UserProfile, error) {
	if err := validateAuthorizationHeader(authorization); err != nil {
		return nil, err
	}

	return uc.auth.GetProfile(ctx, authorization)
}

func (uc *AuthUseCase) Authenticate(ctx context.Context, authorization string) (*domain.UserProfile, error) {
	return uc.GetProfile(ctx, authorization)
}

func (uc *AuthUseCase) UpdateProfile(ctx context.Context, authorization string, input domain.ProfileUpdate) (*domain.UserProfile, error) {
	if err := validateAuthorizationHeader(authorization); err != nil {
		return nil, err
	}

	return uc.auth.UpdateProfile(ctx, authorization, input)
}

func (uc *AuthUseCase) UpdateUserProfile(ctx context.Context, authorization, userID string, input domain.ProfileUpdate) (*domain.UserProfile, error) {
	if err := validateAuthorizationHeader(authorization); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(strings.TrimSpace(userID)); err != nil {
		return nil, fmt.Errorf("%w: userId must be a valid UUID", domain.ErrInvalidInput)
	}

	return uc.auth.UpdateUserProfile(ctx, authorization, userID, input)
}

func (uc *AuthUseCase) GetPublicProfile(ctx context.Context, userID string) (*domain.PublicProfile, error) {
	if _, err := uuid.Parse(strings.TrimSpace(userID)); err != nil {
		return nil, fmt.Errorf("%w: userId must be a valid UUID", domain.ErrInvalidInput)
	}

	return uc.auth.GetPublicProfile(ctx, userID)
}

func (uc *AuthUseCase) CreateInvitation(ctx context.Context, authorization string, input domain.InvitationDraft) (*domain.InvitationLink, error) {
	if err := validateAuthorizationHeader(authorization); err != nil {
		return nil, err
	}
	if err := validateEmail(input.Email); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.FullName) == "" {
		return nil, fmt.Errorf("%w: fullName is required", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(input.Phone) == "" {
		return nil, fmt.Errorf("%w: phone is required", domain.ErrInvalidInput)
	}
	if input.AccessUntil.IsZero() {
		return nil, fmt.Errorf("%w: accessUntil is required", domain.ErrInvalidInput)
	}
	if input.ExpiresAt.IsZero() {
		return nil, fmt.Errorf("%w: expiresAt is required", domain.ErrInvalidInput)
	}
	if !input.ExpiresAt.After(time.Now()) {
		return nil, fmt.Errorf("%w: expiresAt must be in the future", domain.ErrInvalidInput)
	}
	if input.AccessUntil.Before(time.Now().Add(-24 * time.Hour)) {
		return nil, fmt.Errorf("%w: accessUntil must not be in the past", domain.ErrInvalidInput)
	}
	if role, err := normalizeInvitationRole(input.Role); err != nil {
		return nil, err
	} else {
		input.Role = role
	}
	accessUntilEndOfDay := time.Date(
		input.AccessUntil.Year(),
		input.AccessUntil.Month(),
		input.AccessUntil.Day(),
		23, 59, 59, 0,
		input.AccessUntil.Location(),
	)
	if input.ExpiresAt.After(accessUntilEndOfDay) {
		return nil, fmt.Errorf("%w: expiresAt must not be later than accessUntil", domain.ErrInvalidInput)
	}

	token, err := uc.auth.CreateInvitation(ctx, authorization, input)
	if err != nil {
		return nil, err
	}

	baseURL := uc.publicBaseURL
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	invitationURL, err := url.JoinPath(baseURL, "invitations", token)
	if err != nil {
		invitationURL = baseURL + "/invitations/" + token
	}

	return &domain.InvitationLink{
		Token: token,
		URL:   invitationURL,
	}, nil
}

func (uc *AuthUseCase) BlockUser(ctx context.Context, authorization, email string) error {
	if err := validateAuthorizationHeader(authorization); err != nil {
		return err
	}
	if err := validateEmail(email); err != nil {
		return err
	}

	return uc.auth.BlockUser(ctx, authorization, email)
}

func (uc *AuthUseCase) UnblockUser(ctx context.Context, authorization, email string) error {
	if err := validateAuthorizationHeader(authorization); err != nil {
		return err
	}
	if err := validateEmail(email); err != nil {
		return err
	}

	return uc.auth.UnblockUser(ctx, authorization, email)
}

func validateAuthorizationHeader(value string) error {
	if !strings.HasPrefix(strings.TrimSpace(value), "Bearer ") {
		return fmt.Errorf("%w: Authorization header must use Bearer token", domain.ErrInvalidInput)
	}

	return nil
}

func validateEmail(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%w: email is required", domain.ErrInvalidInput)
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("%w: email must be valid", domain.ErrInvalidInput)
	}

	return nil
}

func validatePassword(value string) error {
	if len(strings.TrimSpace(value)) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrInvalidInput)
	}

	return nil
}

func normalizeInvitationRole(value string) (string, error) {
	role := strings.ToLower(strings.TrimSpace(value))
	switch role {
	case "", "psychologist":
		return "psychologist", nil
	default:
		return "", fmt.Errorf("%w: only psychologist invitations are supported", domain.ErrInvalidInput)
	}
}
