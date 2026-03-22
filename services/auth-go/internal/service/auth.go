package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/storage/postgres"
)

type AuthService struct {
	repo *postgres.Storage
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	Expires_in   int64
}

const (
	ErrExpiredToken   string = "token has expired"
	ErrAlreadyUsed    string = "create invitation already used"
	ErrTokenNotExists string = "token doesn`t exists"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountBlocked     = errors.New("account is blocked")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrInvalidSession     = errors.New("invalid session")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
)

func NewAuthService(repo *postgres.Storage) *AuthService {
	return &AuthService{repo: repo}
}

func HashPass(pass string) (string, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), 10)
	if err != nil {
		return "", err
	}
	return string(passHash), nil
}

func (s *AuthService) Authorize(ctx context.Context, requireAdmin bool) (*postgres.AuthState, error) {
	token, err := ExtractToken(ctx)
	if err != nil {
		return nil, err
	}

	claims, err := ValidateAccessToken(token)
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
		return nil, status.Error(codes.PermissionDenied, ErrAccountBlocked.Error())
	case postgres.StatusInactive:
		return nil, status.Error(codes.PermissionDenied, ErrAccountInactive.Error())
	}

	if requireAdmin && authState.Role != postgres.RoleAdmin {
		return nil, status.Error(codes.PermissionDenied, "admin role is required")
	}

	return authState, nil
}

func (s *AuthService) ListPsychologists(ctx context.Context) ([]postgres.DirectoryEntry, error) {
	return s.repo.ListPsychologistDirectory(ctx)
}

func (s *AuthService) CreateInvitation(ctx context.Context, email string, role string, fullName string, phone string, accessUntil time.Time, expiresAt time.Time) (string, error) {
	exists, err := s.repo.UserExistsByEmail(ctx, email)
	if err != nil {
		s.repo.Logger.Error("check invitation user existence failed", slog.Any("error", err), slog.String("email", email))
		return "", err
	}
	if exists {
		return "", ErrUserAlreadyExists
	}

	token, err := s.repo.CreateInvitationUUID(ctx, email, role, fullName, phone, accessUntil, expiresAt)
	if err != nil {
		s.repo.Logger.Error("create invitation failed", slog.Any("error", err), slog.String("email", email))
		return "", err
	}

	return token, nil
}

func (s *AuthService) Register(ctx context.Context, token string, password string) (*TokenDetails, string, error) {
	ok, err := IsAccessToken(ctx, s.repo, token)
	if !ok {
		return nil, "", err
	}

	passwordHash, err := HashPass(password)
	if err != nil {
		return nil, "", err
	}

	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback()

	id, role, err := s.repo.RegisterByInvite(ctx, tx, token, passwordHash)
	if err != nil {
		s.repo.Logger.Error("register by invite failed", slog.Any("error", err))
		return nil, "", err
	}

	tokens, err := GenerateTokensTx(ctx, tx, id, role, s.repo)
	if err != nil {
		return nil, "", err
	}

	if err := tx.Commit(); err != nil {
		return nil, "", err
	}

	return tokens, role, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (*TokenDetails, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	switch user.Status {
	case postgres.StatusBlocked:
		return nil, "", ErrAccountBlocked
	case postgres.StatusInactive:
		return nil, "", ErrAccountInactive
	}

	if ok := CompareHashPass(password, user.PasswordHash); !ok {
		return nil, "", ErrInvalidCredentials
	}

	tokens, err := GenerateTokens(ctx, user.ID, user.Role, s.repo)
	if err != nil {
		return nil, "", err
	}

	return tokens, user.Role, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenDetails, string, error) {
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, "", ErrInvalidSession
	}

	userID, ok := claims["user_id"].(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return nil, "", ErrInvalidSession
	}

	user, err := s.repo.GetUserByRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrInvalidSession
		}
		return nil, "", err
	}

	switch user.Status {
	case postgres.StatusBlocked:
		return nil, "", ErrAccountBlocked
	case postgres.StatusInactive:
		return nil, "", ErrAccountInactive
	}

	newTokens, err := GenerateTokens(ctx, userID, user.Role, s.repo)
	if err != nil {
		return nil, "", err
	}

	return newTokens, user.Role, nil
}

func (s *AuthService) BlockUser(ctx context.Context, email string) error {
	return s.repo.BlockUserByEmail(ctx, email)
}

func (s *AuthService) UnblockUser(ctx context.Context, email string) error {
	return s.repo.UnBlockUserByEmail(ctx, email)
}

func (s *AuthService) GetProfile(ctx context.Context, id string) (*postgres.User, error) {
	return s.repo.GetProfileByID(ctx, id)
}

func (s *AuthService) UpdateProfile(ctx context.Context, id, about, photoURL string) (*postgres.User, error) {
	if err := s.repo.UpdateProfileByID(ctx, id, about, photoURL); err != nil {
		return nil, err
	}

	return s.repo.GetProfileByID(ctx, id)
}

func (s *AuthService) GetPublicProfile(ctx context.Context, id string) (*postgres.PublicProfile, error) {
	return s.repo.GetPublicProfileByID(ctx, id)
}

func IsAccessToken(ctx context.Context, repo *postgres.Storage, token string) (bool, error) {
	expired_at, err := repo.ExpiredToken(ctx, token)
	if err != nil {
		return false, fmt.Errorf(ErrTokenNotExists)
	}
	is_used, err := repo.IsUsedToken(ctx, token)
	if err != nil {
		return false, fmt.Errorf(ErrTokenNotExists)
	}
	if time.Now().After(*expired_at) {
		return false, fmt.Errorf(ErrExpiredToken)
	}
	if is_used {
		return false, fmt.Errorf(ErrAlreadyUsed)
	}
	return true, nil
}

func GenerateTokensTx(ctx context.Context, tx *sql.Tx, id string, role string, repo *postgres.Storage) (*TokenDetails, error) {

	accessTTL := 15 * time.Minute
	accessClaims := jwt.MapClaims{
		"user_id": id,
		"role":    role,
		"exp":     time.Now().Add(accessTTL).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	aToken, err := accessToken.SignedString([]byte(os.Getenv("ACCESS_SECRET_KEY")))
	if err != nil {
		return nil, err
	}

	refreshTTL := 24 * 7 * time.Hour
	refreshClaims := jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(refreshTTL).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rToken, err := refreshToken.SignedString([]byte(os.Getenv("REFRESH_SECRET_KEY")))
	if err != nil {
		return nil, err
	}

	err = repo.UpdateRefreshTokenTx(ctx, tx, rToken, id)
	if err != nil {
		return nil, err
	}

	return &TokenDetails{
		AccessToken:  aToken,
		RefreshToken: rToken,
		Expires_in:   int64(accessTTL.Seconds()),
	}, nil
}

func CompareHashPass(pass string, passHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(pass))
	if err != nil {
		return false
	}
	return true
}

func ValidateRefreshToken(refreshToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("REFRESH_SECRET_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func ValidateAccessToken(accessToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("ACCESS_SECRET_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func ExtractToken(ctx context.Context) (string, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "метаданные отсутствуют")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "заголовок авторизации отсутствует")
	}

	authHeader := values[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "неверный формат заголовка ( нужен bearer )")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func GenerateTokens(ctx context.Context, id string, role string, repo *postgres.Storage) (*TokenDetails, error) {

	accessTTL := 15 * time.Minute
	accessClaims := jwt.MapClaims{
		"user_id": id,
		"role":    role,
		"exp":     time.Now().Add(accessTTL).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	aToken, err := accessToken.SignedString([]byte(os.Getenv("ACCESS_SECRET_KEY")))
	if err != nil {
		return nil, err
	}

	refreshTTL := 24 * 7 * time.Hour
	refreshClaims := jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(refreshTTL).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rToken, err := refreshToken.SignedString([]byte(os.Getenv("REFRESH_SECRET_KEY")))
	if err != nil {
		return nil, err
	}

	err = repo.UpdateRefreshToken(ctx, rToken, id)
	if err != nil {
		return nil, err
	}

	return &TokenDetails{
		AccessToken:  aToken,
		RefreshToken: rToken,
		Expires_in:   int64(accessTTL.Seconds()),
	}, nil
}
