package service

import (
	jwtPkg "DeepSight/internal/util/jwt"
	"context"
	"errors"
	"strings"
	"time"

	"DeepSight/internal/database"
	"DeepSight/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtExpire time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtExpire time.Duration) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtExpire: jwtExpire,
	}
}

func (s *AuthService) Login(username, password string) (string, uint, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", 0, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", 0, errors.New("invalid password")
	}

	token, err := jwtPkg.GenerateToken(user.ID)
	if err != nil {
		return "", 0, errors.New("failed to generate token")
	}

	ctx := context.Background()
	if err := database.SetToken(ctx, user.ID, token, s.jwtExpire); err != nil {
		return "", 0, errors.New("failed to store token")
	}

	return token, user.ID, nil
}

func (s *AuthService) Logout(tokenString string) error {
	claims, err := jwtPkg.ParseToken(tokenString)
	if err != nil {
		return errors.New("invalid token")
	}

	ctx := context.Background()
	return database.DeleteToken(ctx, claims.UserID, tokenString)
}

func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	claims, err := jwtPkg.ParseToken(tokenString)
	if err != nil {
		return 0, errors.New("invalid token")
	}

	ctx := context.Background()

	exists, err := database.TokenExists(ctx, claims.UserID, tokenString)
	if err != nil {
		return 0, errors.New("failed to check token status")
	}
	if !exists {
		return 0, errors.New("token not found or has been revoked")
	}

	return claims.UserID, nil
}

func ExtractToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}

	return strings.TrimSpace(authHeader)
}
