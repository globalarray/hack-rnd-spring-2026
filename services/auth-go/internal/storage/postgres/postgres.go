package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"sourcecraft.dev/benzo/hack-rnd-2026-spring/services/auth-go/internal/config"
)

const (
	RoleAdmin        = "admin"
	RolePsychologist = "psychologist"

	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusBlocked  = "blocked"
)

type Storage struct {
	Config *config.Config
	DB     *sql.DB
	Logger *slog.Logger
}

type User struct {
	ID           string
	Email        string
	FullName     string
	Phone        string
	Role         string
	Status       string
	PasswordHash string
	AccessUntil  time.Time
	PhotoURL     string
	About        string
	RefreshToken string
}

type PublicProfile struct {
	FullName string
	PhotoURL string
	About    string
}

type DirectoryEntry struct {
	ID              string
	Email           string
	FullName        string
	Phone           string
	Role            string
	Status          string
	AccessUntil     time.Time
	ExpiresAt       time.Time
	InvitationToken string
}

type AuthState struct {
	ID          string
	Role        string
	Status      string
	AccessUntil time.Time
}

func NewStorage(cfg *config.Config, logger *slog.Logger) (*Storage, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	storage := &Storage{
		Config: cfg,
		DB:     db,
		Logger: logger,
	}

	if err := storage.runMigrations(); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := storage.syncExpiredUsers(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := storage.ensureBootstrapAdmin(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return storage, nil
}

func (s *Storage) runMigrations() error {
	content, err := os.ReadFile(s.Config.MigrationsPath)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	if _, err := s.DB.Exec(string(content)); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case RoleAdmin:
		return RoleAdmin
	default:
		return RolePsychologist
	}
}

func parseAccessUntil(raw string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, fmt.Errorf("parse access until: %w", err)
	}

	return parsed, nil
}

func (s *Storage) ensureBootstrapAdmin(ctx context.Context) error {
	var existingID string
	err := s.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, s.Config.BootstrapAdmin.Email).Scan(&existingID)
	if err == nil {
		s.Logger.Info("bootstrap admin already exists", slog.String("email", s.Config.BootstrapAdmin.Email))
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check bootstrap admin: %w", err)
	}

	accessUntil, err := parseAccessUntil(s.Config.BootstrapAdmin.AccessUntil)
	if err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(s.Config.BootstrapAdmin.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap admin password: %w", err)
	}

	role := normalizeRole(s.Config.BootstrapAdmin.Role)

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO users (email, full_name, password_hash, phone, role, status, access_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		s.Config.BootstrapAdmin.Email,
		s.Config.BootstrapAdmin.FullName,
		string(passwordHash),
		s.Config.BootstrapAdmin.Phone,
		role,
		StatusActive,
		accessUntil,
	)
	if err != nil {
		return fmt.Errorf("insert bootstrap admin: %w", err)
	}

	s.Logger.Info("bootstrap admin created", slog.String("email", s.Config.BootstrapAdmin.Email))
	return nil
}

func (s *Storage) syncExpiredUsers(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE users
		SET status = $1, refresh_token = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE status <> $2
		  AND access_until < CURRENT_DATE
	`, StatusInactive, StatusBlocked)
	if err != nil {
		return fmt.Errorf("sync expired users: %w", err)
	}

	return nil
}

func (s *Storage) CreateInvitationUUID(ctx context.Context, email string, role string, fullName string, phone string, accessUntil time.Time, expiresAt time.Time) (string, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM psychologist_invitations
		WHERE email = $1
		  AND is_used = FALSE
	`, email); err != nil {
		return "", err
	}

	query := `
		INSERT INTO psychologist_invitations (email, role, full_name, phone, access_until, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING token
	`

	var token string
	if err := tx.QueryRowContext(ctx, query, email, normalizeRole(role), fullName, phone, accessUntil, expiresAt).Scan(&token); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Storage) ExpiredToken(ctx context.Context, token string) (*time.Time, error) {
	var expiresAt time.Time
	err := s.DB.QueryRowContext(ctx, `SELECT expires_at FROM psychologist_invitations WHERE token = $1`, token).Scan(&expiresAt)
	if err != nil {
		return nil, err
	}

	return &expiresAt, nil
}

func (s *Storage) IsUsedToken(ctx context.Context, token string) (bool, error) {
	var isUsed bool
	err := s.DB.QueryRowContext(ctx, `SELECT is_used FROM psychologist_invitations WHERE token = $1`, token).Scan(&isUsed)
	if err != nil {
		return false, err
	}

	return isUsed, nil
}

func (s *Storage) RegisterByInvite(ctx context.Context, tx *sql.Tx, token string, passHash string) (id, role string, err error) {
	var (
		email       string
		fullName    string
		phone       string
		accessUntil time.Time
		inviteRole  string
	)

	selectQuery := `
		SELECT email, full_name, phone, role, access_until
		FROM psychologist_invitations
		WHERE token = $1
		  AND is_used = FALSE
		  AND expires_at > CURRENT_TIMESTAMP
	`

	if err := tx.QueryRowContext(ctx, selectQuery, token).Scan(&email, &fullName, &phone, &inviteRole, &accessUntil); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("invitation not found")
		}
		return "", "", err
	}

	status := StatusActive
	if accessUntil.Before(time.Now()) {
		status = StatusInactive
	}

	insertQuery := `
		INSERT INTO users (email, full_name, password_hash, phone, role, status, access_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	if err := tx.QueryRowContext(ctx, insertQuery, email, fullName, passHash, phone, normalizeRole(inviteRole), status, accessUntil).Scan(&id); err != nil {
		return "", "", err
	}

	updateInviteQuery := `
		UPDATE psychologist_invitations
		SET is_used = TRUE, used_at = CURRENT_TIMESTAMP
		WHERE token = $1
	`

	if _, err := tx.ExecContext(ctx, updateInviteQuery, token); err != nil {
		return "", "", err
	}

	return id, normalizeRole(inviteRole), nil
}

func (s *Storage) UpdateRefreshTokenTx(ctx context.Context, tx *sql.Tx, refreshToken string, id string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE users
		SET refresh_token = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, refreshToken, id)
	return err
}

func (s *Storage) UpdateRefreshToken(ctx context.Context, refreshToken string, id string) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE users
		SET refresh_token = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, refreshToken, id)
	return err
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if err := s.syncExpiredUsers(ctx); err != nil {
		return nil, err
	}

	row := &User{}
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, email, full_name, phone, role, status, password_hash, access_until, COALESCE(photo_url, ''), COALESCE(about, ''), COALESCE(refresh_token, '')
		FROM users
		WHERE email = $1
	`, email).Scan(
		&row.ID,
		&row.Email,
		&row.FullName,
		&row.Phone,
		&row.Role,
		&row.Status,
		&row.PasswordHash,
		&row.AccessUntil,
		&row.PhotoURL,
		&row.About,
		&row.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	return row, nil
}

func (s *Storage) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE email = $1
		)
	`, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Storage) GetUserByRefreshToken(ctx context.Context, id, refreshToken string) (*User, error) {
	if err := s.syncExpiredUsers(ctx); err != nil {
		return nil, err
	}

	row := &User{}
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, email, full_name, phone, role, status, password_hash, access_until, COALESCE(photo_url, ''), COALESCE(about, ''), COALESCE(refresh_token, '')
		FROM users
		WHERE refresh_token = $1 AND id = $2
	`, refreshToken, id).Scan(
		&row.ID,
		&row.Email,
		&row.FullName,
		&row.Phone,
		&row.Role,
		&row.Status,
		&row.PasswordHash,
		&row.AccessUntil,
		&row.PhotoURL,
		&row.About,
		&row.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	return row, nil
}

func (s *Storage) GetUserAuthStateByID(ctx context.Context, id string) (*AuthState, error) {
	if err := s.syncExpiredUsers(ctx); err != nil {
		return nil, err
	}

	state := &AuthState{}
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, role, status, access_until
		FROM users
		WHERE id = $1
	`, id).Scan(&state.ID, &state.Role, &state.Status, &state.AccessUntil)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *Storage) BlockUserByEmail(ctx context.Context, email string) error {
	result, err := s.DB.ExecContext(ctx, `
		UPDATE users
		SET status = $1, refresh_token = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE email = $2
	`, StatusBlocked, email)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Storage) UnBlockUserByEmail(ctx context.Context, email string) error {
	result, err := s.DB.ExecContext(ctx, `
		UPDATE users
		SET status = CASE WHEN access_until < CURRENT_DATE THEN $1 ELSE $2 END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE email = $3
	`, StatusInactive, StatusActive, email)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Storage) GetProfileByID(ctx context.Context, id string) (*User, error) {
	if err := s.syncExpiredUsers(ctx); err != nil {
		return nil, err
	}

	row := &User{}
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, email, full_name, phone, role, status, access_until, COALESCE(photo_url, ''), COALESCE(about, '')
		FROM users
		WHERE id = $1
	`, id).Scan(
		&row.ID,
		&row.Email,
		&row.FullName,
		&row.Phone,
		&row.Role,
		&row.Status,
		&row.AccessUntil,
		&row.PhotoURL,
		&row.About,
	)
	if err != nil {
		return nil, err
	}

	return row, nil
}

func (s *Storage) UpdateProfileByID(ctx context.Context, id, about, photoURL string) error {
	result, err := s.DB.ExecContext(ctx, `
		UPDATE users
		SET about = $1, photo_url = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, about, photoURL, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Storage) GetPublicProfileByID(ctx context.Context, id string) (*PublicProfile, error) {
	row := &PublicProfile{}
	err := s.DB.QueryRowContext(ctx, `
		SELECT full_name, COALESCE(photo_url, ''), COALESCE(about, '')
		FROM users
		WHERE id = $1
	`, id).Scan(&row.FullName, &row.PhotoURL, &row.About)
	if err != nil {
		return nil, err
	}

	return row, nil
}

func (s *Storage) ListPsychologistDirectory(ctx context.Context) ([]DirectoryEntry, error) {
	if err := s.syncExpiredUsers(ctx); err != nil {
		return nil, err
	}

	items := make([]DirectoryEntry, 0)

	userRows, err := s.DB.QueryContext(ctx, `
		SELECT id, email, full_name, phone, role, status, access_until
		FROM users
		WHERE role = $1
	`, RolePsychologist)
	if err != nil {
		return nil, err
	}
	defer userRows.Close()

	for userRows.Next() {
		var item DirectoryEntry
		if err := userRows.Scan(
			&item.ID,
			&item.Email,
			&item.FullName,
			&item.Phone,
			&item.Role,
			&item.Status,
			&item.AccessUntil,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}
	if err := userRows.Err(); err != nil {
		return nil, err
	}

	invitationRows, err := s.DB.QueryContext(ctx, `
		SELECT email, full_name, phone, role, access_until, expires_at, token
		FROM psychologist_invitations
		WHERE is_used = FALSE
	`)
	if err != nil {
		return nil, err
	}
	defer invitationRows.Close()

	for invitationRows.Next() {
		var item DirectoryEntry
		if err := invitationRows.Scan(
			&item.Email,
			&item.FullName,
			&item.Phone,
			&item.Role,
			&item.AccessUntil,
			&item.ExpiresAt,
			&item.InvitationToken,
		); err != nil {
			return nil, err
		}

		item.Status = "pending"
		items = append(items, item)
	}
	if err := invitationRows.Err(); err != nil {
		return nil, err
	}

	sort.SliceStable(items, func(i, j int) bool {
		leftName := strings.ToLower(strings.TrimSpace(items[i].FullName))
		rightName := strings.ToLower(strings.TrimSpace(items[j].FullName))
		if leftName == rightName {
			return strings.ToLower(strings.TrimSpace(items[i].Email)) < strings.ToLower(strings.TrimSpace(items[j].Email))
		}
		return leftName < rightName
	})

	return items, nil
}
