// Package application implements the business logic layer (use cases) of the API Gateway.
package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"prodlich/internal/domain/entity"
	"prodlich/internal/domain/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when login/registration data is invalid.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidName        = errors.New("name is required (1-100 characters)")
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// AuthUseCase handles regular user registration, login and JWT issuance.
type AuthUseCase struct {
	userRepo  repository.UserRepo
	jwtSecret []byte
	bcryptCost int
	tokenTTL   time.Duration
}

// NewAuthUseCase wires up an AuthUseCase. bcryptCost defaults to 12 and
// tokenTTL to 24h when zero values are passed.
func NewAuthUseCase(userRepo repository.UserRepo, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		jwtSecret:  []byte(jwtSecret),
		bcryptCost: 12,
		tokenTTL:   24 * time.Hour,
	}
}

// Register creates a new user and returns a signed JWT.
func (uc *AuthUseCase) Register(ctx context.Context, name, email, password string) (string, error) {
	if err := validateCredentials(name, email, password); err != nil {
		return "", err
	}
	email = strings.ToLower(strings.TrimSpace(email))

	if existing, _ := uc.userRepo.GetByEmail(ctx, email); existing != nil {
		return "", ErrEmailExists
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), uc.bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	user := entity.User{
		Name:      strings.TrimSpace(name),
		Email:     email,
		Password:  string(hashed),
		CreatedAt: time.Now().UTC(),
	}
	id, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	return uc.issueToken(id, email)
}

// Login authenticates a user and returns a signed JWT.
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return "", ErrInvalidCredentials
	}
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("lookup user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return uc.issueToken(user.ID, user.Email)
}

// ValidateToken parses tokenString, verifies the signature and returns the user.
func (uc *AuthUseCase) ValidateToken(ctx context.Context, tokenString string) (*entity.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return uc.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidCredentials
	}
	idF, ok := claims["user_id"].(float64)
	if !ok {
		return nil, ErrInvalidCredentials
	}
	user, err := uc.userRepo.GetByID(ctx, int64(idF))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (uc *AuthUseCase) issueToken(userID int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(uc.tokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.jwtSecret)
}

func validateCredentials(name, email, password string) error {
	if l := len(strings.TrimSpace(name)); l == 0 || l > 100 {
		return ErrInvalidName
	}
	if !emailRe.MatchString(strings.TrimSpace(email)) {
		return ErrInvalidEmail
	}
	if len(password) < 8 {
		return ErrWeakPassword
	}
	return nil
}
