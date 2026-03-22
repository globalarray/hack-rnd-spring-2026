package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func HashPass(pass string) (string, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), 10)
	if err != nil {
		return "", err
	}
	return string(passHash), nil
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
