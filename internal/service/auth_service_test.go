package service

import (
	"errors"
	"testing"
	"time"

	"github.com/roychanmeliaz/btechdevcases/internal/models"
	"github.com/roychanmeliaz/btechdevcases/internal/repository"
	customjwt "github.com/roychanmeliaz/btechdevcases/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Auto migrate tables
	err = db.AutoMigrate(&models.User{}, &models.Wallet{}, &models.Transaction{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}


func TestAuthService_Register(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	jwtManager := customjwt.NewManager("test-secret", 24*time.Hour)

	authService := NewAuthService(userRepo, walletRepo, jwtManager, db)

	t.Run("successful registration", func(t *testing.T) {
		user, err := authService.Register("test@example.com", "password123", "password123")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if user == nil {
			t.Error("expected user, got nil")
		}
		if user.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", user.Email)
		}

		// Verify wallet was created
		wallet, err := walletRepo.FindByUserID(user.ID)
		if err != nil {
			t.Errorf("expected wallet to be created, got error: %v", err)
		}
		if wallet.Balance != 1000.0 {
			t.Errorf("expected initial balance 1000, got %f", wallet.Balance)
		}
	})

	t.Run("password mismatch", func(t *testing.T) {
		_, err := authService.Register("test2@example.com", "password123", "password456")
		if !errors.Is(err, ErrPasswordMismatch) {
			t.Errorf("expected ErrPasswordMismatch, got %v", err)
		}
	})

	t.Run("weak password", func(t *testing.T) {
		_, err := authService.Register("test3@example.com", "pass", "pass")
		if !errors.Is(err, ErrWeakPassword) {
			t.Errorf("expected ErrWeakPassword, got %v", err)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := authService.Register("test@example.com", "password123", "password123")
		if !errors.Is(err, ErrEmailExists) {
			t.Errorf("expected ErrEmailExists, got %v", err)
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	jwtManager := customjwt.NewManager("test-secret", 24*time.Hour)

	authService := NewAuthService(userRepo, walletRepo, jwtManager, db)

	// Create a test user
	email := "login@example.com"
	password := "password123"
	_, err := authService.Register(email, password, password)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Run("successful login", func(t *testing.T) {
		token, user, err := authService.Login(email, password)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if token == "" {
			t.Error("expected token, got empty string")
		}
		if user == nil {
			t.Error("expected user, got nil")
		}
		if user.Email != email {
			t.Errorf("expected email %s, got %s", email, user.Email)
		}

		// Verify token can be validated
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			t.Errorf("expected valid token, got error: %v", err)
		}
		if claims.Email != email {
			t.Errorf("expected email %s in token, got %s", email, claims.Email)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		_, _, err := authService.Login("nonexistent@example.com", password)
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		_, _, err := authService.Login(email, "wrongpassword")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"
	
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Verify correct password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		t.Error("expected password to match")
	}

	// Verify incorrect password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	if err == nil {
		t.Error("expected password to not match")
	}
}
