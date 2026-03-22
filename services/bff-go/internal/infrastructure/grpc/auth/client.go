package auth

import (
	"context"
	"strings"

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

	return mapTokens(resp), nil
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	resp, err := c.client.RefreshToken(ctx, &authpb.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, err
	}

	return mapTokens(resp), nil
}

func (c *Client) Register(ctx context.Context, token, password string) (*domain.AuthTokens, error) {
	resp, err := c.client.Register(ctx, &authpb.RegisterRequest{
		Token:    token,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	return mapTokens(resp), nil
}

func (c *Client) GetProfile(ctx context.Context, authorization string) (*domain.UserProfile, error) {
	resp, err := c.client.GetProfile(withAuthorization(ctx, authorization), &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return mapProfile(resp), nil
}

func (c *Client) UpdateProfile(ctx context.Context, authorization string, input domain.ProfileUpdate) (*domain.UserProfile, error) {
	resp, err := c.client.UpdateProfile(withAuthorization(ctx, authorization), &authpb.UpdateProfileRequest{
		PhotoUrl: input.PhotoURL,
		About:    input.About,
	})
	if err != nil {
		return nil, err
	}

	return mapProfile(resp), nil
}

func (c *Client) GetPublicProfile(ctx context.Context, userID string) (*domain.PublicProfile, error) {
	resp, err := c.client.GetPublicProfile(ctx, &authpb.PublicProfileRequest{UserId: userID})
	if err != nil {
		return nil, err
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

	return resp.GetToken(), nil
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

func mapTokens(resp *authpb.TokenResponse) *domain.AuthTokens {
	if resp == nil {
		return nil
	}

	return &domain.AuthTokens{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
		ExpiresIn:    resp.GetExpiresIn(),
		Role:         resp.GetRole(),
	}
}

func mapProfile(resp *authpb.ProfileResponse) *domain.UserProfile {
	if resp == nil {
		return nil
	}

	profile := &domain.UserProfile{
		ID:       resp.GetId(),
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

	return profile
}
