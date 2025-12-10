package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/brady1408/dnd/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrKeyTaken           = errors.New("SSH key already registered")
)

// Service handles authentication
type Service struct {
	queries *db.Queries
}

// NewService creates a new auth service
func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a password against a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// NormalizePublicKey extracts and normalizes the key data
func NormalizePublicKey(key ssh.PublicKey) string {
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
}

// isValidUUID checks if a pgtype.UUID is valid (not null/empty)
func isValidUUID(id pgtype.UUID) bool {
	return id.Valid
}

// RegisterWithPassword registers a new user with email and password
func (s *Service) RegisterWithPassword(ctx context.Context, email, password string) (*db.User, error) {
	// Check if email already exists
	existing, err := s.queries.GetUserByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err == nil && isValidUUID(existing.ID) {
		return nil, ErrEmailTaken
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err := s.queries.CreateUserWithPassword(ctx, db.CreateUserWithPasswordParams{
		Email:        pgtype.Text{String: email, Valid: true},
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// RegisterWithPublicKey registers a new user with SSH public key
func (s *Service) RegisterWithPublicKey(ctx context.Context, key ssh.PublicKey) (*db.User, error) {
	keyStr := NormalizePublicKey(key)

	// Check if key already exists
	existing, err := s.queries.GetUserByPublicKey(ctx, pgtype.Text{String: keyStr, Valid: true})
	if err == nil && isValidUUID(existing.ID) {
		return nil, ErrKeyTaken
	}

	user, err := s.queries.CreateUserWithPublicKey(ctx, pgtype.Text{String: keyStr, Valid: true})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// LoginWithPassword authenticates a user with email and password
func (s *Service) LoginWithPassword(ctx context.Context, email, password string) (*db.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.PasswordHash.Valid || !CheckPassword(password, user.PasswordHash.String) {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

// LoginWithPublicKey authenticates a user with SSH public key
func (s *Service) LoginWithPublicKey(ctx context.Context, key ssh.PublicKey) (*db.User, error) {
	keyStr := NormalizePublicKey(key)
	user, err := s.queries.GetUserByPublicKey(ctx, pgtype.Text{String: keyStr, Valid: true})
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, id pgtype.UUID) (*db.User, error) {
	user, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

// LinkPublicKey links an SSH public key to an existing user
func (s *Service) LinkPublicKey(ctx context.Context, userID pgtype.UUID, key ssh.PublicKey) error {
	keyStr := NormalizePublicKey(key)

	// Check if key is already taken by another user
	existing, err := s.queries.GetUserByPublicKey(ctx, pgtype.Text{String: keyStr, Valid: true})
	if err == nil && isValidUUID(existing.ID) && existing.ID != userID {
		return ErrKeyTaken
	}

	_, err = s.queries.UpdateUserPublicKey(ctx, db.UpdateUserPublicKeyParams{
		ID:        userID,
		PublicKey: pgtype.Text{String: keyStr, Valid: true},
	})
	return err
}

// UpdateEmail updates a user's email
func (s *Service) UpdateEmail(ctx context.Context, userID pgtype.UUID, email string) error {
	// Check if email is taken
	existing, err := s.queries.GetUserByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err == nil && isValidUUID(existing.ID) && existing.ID != userID {
		return ErrEmailTaken
	}

	_, err = s.queries.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
		ID:    userID,
		Email: pgtype.Text{String: email, Valid: true},
	})
	return err
}

// UpdatePassword updates a user's password
func (s *Service) UpdatePassword(ctx context.Context, userID pgtype.UUID, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	_, err = s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	return err
}
