package application

import (
    "context"
    "errors"
    "fmt"
    "prodlich/internal/domain/entity"
    "prodlich/internal/domain/repository"
    authpkg "prodlich/internal/infrastructure/auth"

    "golang.org/x/crypto/bcrypt"
)

type AdminTokenService interface {
    GenerateAdminToken(userID int, username, role string) (string, error)
    ValidateAdminToken(tokenString string) (*authpkg.AdminClaims, error)
}

type AdminAuthUseCase struct {
	adminRepo repository.AdminRepo
	tokenSvc  AdminTokenService
}

func NewAdminAuthUseCase(adminRepo repository.AdminRepo, tokenSvc AdminTokenService) *AdminAuthUseCase {
	return &AdminAuthUseCase{
		adminRepo: adminRepo,
		tokenSvc:  tokenSvc,
	}
}

func (uc *AdminAuthUseCase) Login(ctx context.Context, username, password string) (string, error) {
	admin, err := uc.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", errors.New("invalid credentials")
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	go func() {
		_ = uc.adminRepo.UpdateLastLogin(context.Background(), admin.ID)
	}()

	token, err := uc.tokenSvc.GenerateAdminToken(admin.ID, admin.Username, admin.Role)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

func (uc *AdminAuthUseCase) ValidateToken(ctx context.Context, tokenString string) (*entity.AdminUser, error) {
    claims, err := uc.tokenSvc.ValidateAdminToken(tokenString)
    if err != nil {
        return nil, errors.New("invalid token")
    }
    admin, err := uc.adminRepo.GetByID(ctx, claims.UserID)
    if err != nil {
        return nil, errors.New("admin not found")
    }
    return &admin, nil
}
