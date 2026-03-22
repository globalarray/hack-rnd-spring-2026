package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	authpb "sourcecraft.dev/benzo/bff/internal/clients/grpc/authpb"
	"sourcecraft.dev/benzo/bff/internal/domain"
)

type Client struct {
	client authpb.AuthServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{client: authpb.NewAuthServiceClient(conn)}
}

func (c *Client) Login(ctx context.Context, email, password string) (*domain.AuthTokens, error) {
	resp, err := c.client.Login(ctx, &authpb.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return mapTokens(resp)
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	resp, err := c.client.RefreshToken(ctx, &authpb.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, err
	}

	return mapTokens(resp)
}

func (c *Client) Register(ctx context.Context, token, password string) (*domain.AuthTokens, error) {
	resp, err := c.client.Register(ctx, &authpb.RegisterRequest{
		Token:    token,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return mapTokens(resp)
}

func (c *Client) GetProfile(ctx context.Context, authorization string) (*domain.UserProfile, error) {
	resp, err := c.client.GetProfile(withAuthorization(ctx, authorization), &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return mapProfile(resp)
}

func (c *Client) UpdateProfile(ctx context.Context, authorization string, input domain.ProfileUpdate) (*domain.UserProfile, error) {
	resp, err := c.client.UpdateProfile(withAuthorization(ctx, authorization), &authpb.UpdateProfileRequest{
		PhotoUrl: input.PhotoURL,
		About:    input.About,
	})
	if err != nil {
		return nil, err
	}

	return mapProfile(resp)
}

func (c *Client) UpdateUserProfile(ctx context.Context, authorization, userID string, input domain.ProfileUpdate) (*domain.UserProfile, error) {
	resp, err := c.client.UpdateUserProfile(withAuthorization(ctx, authorization), &authpb.UpdateUserProfileRequest{
		UserId:   userID,
		PhotoUrl: input.PhotoURL,
		About:    input.About,
	})
	if err != nil {
		return nil, err
	}

	return mapProfile(resp)
}

func (c *Client) GetPublicProfile(ctx context.Context, userID string) (*domain.PublicProfile, error) {
	resp, err := c.client.GetPublicProfile(ctx, &authpb.PublicProfileRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(resp.GetFullName()) == "" {
		return nil, fmt.Errorf("%w: auth public profile is missing fullName", domain.ErrUpstreamResponse)
	}

	return &domain.PublicProfile{
		FullName: resp.GetFullName(),
		PhotoURL: resp.GetPhotoUrl(),
		About:    resp.GetAbout(),
	}, nil
}

func (c *Client) CreateInvitation(ctx context.Context, authorization string, input domain.InvitationDraft) (string, error) {
	resp, err := c.client.CreateInvitation(withAuthorization(ctx, authorization), &authpb.CreateInvitationRequest{
		FullName:    input.FullName,
		Number:      input.Phone,
		Email:       input.Email,
		Role:        strings.ToLower(strings.TrimSpace(input.Role)),
		AccessUntil: timestamppb.New(input.AccessUntil),
		ExpiresAt:   timestamppb.New(input.ExpiresAt),
	})
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(resp.GetToken())
	if _, err := uuid.Parse(token); err != nil {
		return "", fmt.Errorf("%w: auth invitation token is invalid", domain.ErrUpstreamResponse)
	}

	return token, nil
}

func (c *Client) BlockUser(ctx context.Context, authorization, email string) error {
	_, err := c.client.BlockUser(withAuthorization(ctx, authorization), &authpb.BlockUserRequest{Email: email})
	return err
}

func (c *Client) UnblockUser(ctx context.Context, authorization, email string) error {
	_, err := c.client.UnBlockUser(withAuthorization(ctx, authorization), &authpb.UnBlockUserRequest{Email: email})
	return err
}

func withAuthorization(ctx context.Context, authorization string) context.Context {
	md := metadata.Pairs("authorization", authorization)
	return metadata.NewOutgoingContext(ctx, md)
}

func mapTokens(resp *authpb.TokenResponse) (*domain.AuthTokens, error) {
	if resp == nil {
		return nil, fmt.Errorf("%w: auth token response is empty", domain.ErrUpstreamResponse)
	}

	if strings.TrimSpace(resp.GetAccessToken()) == "" || strings.TrimSpace(resp.GetRefreshToken()) == "" {
		return nil, fmt.Errorf("%w: auth token response is missing tokens", domain.ErrUpstreamResponse)
	}

	if strings.TrimSpace(resp.GetRole()) == "" {
		return nil, fmt.Errorf("%w: auth token response is missing role", domain.ErrUpstreamResponse)
	}

	return &domain.AuthTokens{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
		ExpiresIn:    resp.GetExpiresIn(),
		Role:         resp.GetRole(),
	}, nil
}

func mapProfile(resp *authpb.ProfileResponse) (*domain.UserProfile, error) {
	if resp == nil {
		return nil, fmt.Errorf("%w: auth profile response is empty", domain.ErrUpstreamResponse)
	}

	id := strings.TrimSpace(resp.GetId())
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("%w: auth profile id is invalid", domain.ErrUpstreamResponse)
	}

	if strings.TrimSpace(resp.GetEmail()) == "" {
		return nil, fmt.Errorf("%w: auth profile email is missing", domain.ErrUpstreamResponse)
	}

	if strings.TrimSpace(resp.GetRole()) == "" {
		return nil, fmt.Errorf("%w: auth profile role is missing", domain.ErrUpstreamResponse)
	}

	profile := &domain.UserProfile{
		ID:       id,
		Email:    resp.GetEmail(),
		FullName: resp.GetFullName(),
		Phone:    resp.GetPhone(),
		Role:     resp.GetRole(),
		Status:   resp.GetStatus(),
		PhotoURL: resp.GetPhotoUrl(),
		About:    resp.GetAbout(),
	}

	if accessUntil := resp.GetAccessUntil(); accessUntil != nil {
		profile.AccessUntil = accessUntil.AsTime()
	}

	return profile, nil
}
