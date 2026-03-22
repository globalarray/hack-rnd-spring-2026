package domain

import "time"

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	Role         string
}

type UserProfile struct {
	ID          string
	Email       string
	FullName    string
	Phone       string
	Role        string
	Status      string
	PhotoURL    string
	About       string
	AccessUntil time.Time
}

type ProfileUpdate struct {
	PhotoURL string
	About    string
}

type PublicProfile struct {
	FullName string
	PhotoURL string
	About    string
}

type InvitationDraft struct {
	FullName    string
	Phone       string
	Email       string
	Role        string
	AccessUntil time.Time
	ExpiresAt   time.Time
}

type InvitationLink struct {
	Token string
	URL   string
}
