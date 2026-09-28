package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"shalom/internal/middleware"
	"shalom/internal/models"
	"shalom/internal/repository"
)

type AuthService struct {
	rdb      *redis.Client
	auth     *middleware.AuthMiddleware
	userRepo *repository.UserRepository
}

func NewAuthService(rdb *redis.Client, auth *middleware.AuthMiddleware, userRepo *repository.UserRepository) *AuthService {
	return &AuthService{rdb: rdb, auth: auth, userRepo: userRepo}
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

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		PasswordHash: string(hash),
		Status:       models.UserOnline,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("email or username already exists")
	}

	// Create default settings
	settings := &models.UserSettings{
		UserID:              user.ID,
		NotificationsEnabled: true,
		SoundEnabled:         true,
		DefaultBandwidth:     models.ModeStandard,
		Language:             "en",
		Theme:                "dark",
		AutoReconnect:        true,
		ShowOnlineStatus:     true,
		AllowDirectMessages:  true,
		CameraDefaultOn:      false,
		MicDefaultMuted:      true,
	}
	s.userRepo.UpsertSettings(ctx, settings)

	token, err := s.auth.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	s.rdb.Set(ctx, "session:"+user.ID.String(), "online", 72*time.Hour)

	return &AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	s.userRepo.UpdateStatus(ctx, user.ID, models.UserOnline)
	s.rdb.Set(ctx, "session:"+user.ID.String(), "online", 72*time.Hour)

	token, err := s.auth.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	user.Status = models.UserOnline
	return &AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Logout(ctx context.Context, token string, userID uuid.UUID) error {
	s.auth.BlacklistToken(token)
	s.userRepo.UpdateStatus(ctx, userID, models.UserOffline)
	s.rdb.Del(ctx, "session:"+userID.String())
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL, bio, phone string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if displayName != "" {
		user.DisplayName = displayName
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if bio != "" {
		user.Bio = bio
	}
	if phone != "" {
		user.Phone = phone
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	return s.userRepo.Search(ctx, query, 20)
}

func (s *AuthService) GetSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	return s.userRepo.GetSettings(ctx, userID)
}

func (s *AuthService) UpdateSettings(ctx context.Context, settings *models.UserSettings) error {
	return s.userRepo.UpsertSettings(ctx, settings)
}
