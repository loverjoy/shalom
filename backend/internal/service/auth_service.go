package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"shalom/internal/middleware"
	"shalom/internal/models"
)

type AuthService struct {
	db    *pgxpool.Pool
	rdb   *redis.Client
	auth  *middleware.AuthMiddleware
}

func NewAuthService(db *pgxpool.Pool, rdb *redis.Client, auth *middleware.AuthMiddleware) *AuthService {
	return &AuthService{db: db, rdb: rdb, auth: auth}
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=30"`
	DisplayName string `json:"display_name" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.db.QueryRow(ctx,
		`INSERT INTO users (id, email, username, display_name, password_hash, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, 'online', NOW(), NOW())
		 RETURNING id, email, username, display_name, avatar_url, status, last_seen, created_at, updated_at`,
		uuid.New(), req.Email, req.Username, req.DisplayName, string(hash),
	).Scan(&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.AvatarURL, &user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errors.New("email or username already exists")
	}

	token, err := s.auth.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	// Store session in Redis for presence
	s.rdb.Set(ctx, "session:"+user.ID.String(), "online", 72*time.Hour)

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	var user models.User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, username, display_name, password_hash, avatar_url, status, last_seen, created_at, updated_at
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash, &user.AvatarURL, &user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.auth.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	// Update online status
	s.db.Exec(ctx, `UPDATE users SET status = 'online', last_seen = NOW() WHERE id = $1`, user.ID)
	s.rdb.Set(ctx, "session:"+user.ID.String(), "online", 72*time.Hour)

	user.Status = "online"
	return &AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	s.auth.BlacklistToken(token)
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx,
		`SELECT id, email, username, display_name, avatar_url, status, last_seen, created_at, updated_at
		 FROM users WHERE id = $1`, userID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.AvatarURL, &user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func generateMeetingCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)[:8]
}
