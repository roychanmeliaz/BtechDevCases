package service

import (
	"errors"
	"fmt"

	"github.com/roychanmeliaz/btechdevcases/internal/models"
	"github.com/roychanmeliaz/btechdevcases/internal/repository"
	customjwt "github.com/roychanmeliaz/btechdevcases/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrPasswordMismatch  = errors.New("passwords do not match")
	ErrWeakPassword      = errors.New("password must be at least 8 characters")
)

type AuthService interface {
	Register(email, password, confirmPassword string) (*models.User, error)
	Login(email, password string) (string, *models.User, error)
}

type authService struct {
	userRepo   repository.UserRepository
	walletRepo repository.WalletRepository
	jwtManager *customjwt.Manager
	db         *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	walletRepo repository.WalletRepository,
	jwtManager *customjwt.Manager,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepo:   userRepo,
		walletRepo: walletRepo,
		jwtManager: jwtManager,
		db:         db,
	}
}

func (s *authService) Register(email, password, confirmPassword string) (*models.User, error) {
	// Validate password match
	if password != confirmPassword {
		return nil, ErrPasswordMismatch
	}

	// Validate password strength
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking email: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user and wallet in a transaction
	var user *models.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create user
		user = &models.User{
			Email:    email,
			Password: string(hashedPassword),
		}
		if err := s.userRepo.Create(user); err != nil {
			return fmt.Errorf("error creating user: %w", err)
		}

		// Create wallet for user with initial balance of 1000
		wallet := &models.Wallet{
			UserID:  user.ID,
			Balance: 1000.0, // Initial balance
		}
		if err := s.walletRepo.Create(wallet); err != nil {
			return fmt.Errorf("error creating wallet: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, *models.User, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("error finding user: %w", err)
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", nil, fmt.Errorf("error generating token: %w", err)
	}

	return token, user, nil
}
