package jwt

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	manager := NewManager("test-secret", 24*time.Hour)

	t.Run("generate valid token", func(t *testing.T) {
		token, err := manager.GenerateToken(1, "test@example.com")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if token == "" {
			t.Error("expected token, got empty string")
		}
	})

	t.Run("token contains user info", func(t *testing.T) {
		userID := uint(123)
		email := "user@example.com"

		token, err := manager.GenerateToken(userID, email)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		claims, err := manager.ValidateToken(token)
		if err != nil {
			t.Errorf("failed to validate token: %v", err)
		}
		if claims.UserID != userID {
			t.Errorf("expected user ID %d, got %d", userID, claims.UserID)
		}
		if claims.Email != email {
			t.Errorf("expected email %s, got %s", email, claims.Email)
		}
	})
}

func TestJWTManager_ValidateToken(t *testing.T) {
	manager := NewManager("test-secret", 24*time.Hour)

	t.Run("validate valid token", func(t *testing.T) {
		token, _ := manager.GenerateToken(1, "test@example.com")

		claims, err := manager.ValidateToken(token)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if claims == nil {
			t.Error("expected claims, got nil")
		}
	})

	t.Run("reject invalid token", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid-token")
		if err == nil {
			t.Error("expected error for invalid token, got nil")
		}
		if err != ErrInvalidToken {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("reject token with wrong secret", func(t *testing.T) {
		wrongManager := NewManager("wrong-secret", 24*time.Hour)
		token, _ := manager.GenerateToken(1, "test@example.com")

		_, err := wrongManager.ValidateToken(token)
		if err == nil {
			t.Error("expected error for token with wrong secret, got nil")
		}
	})

	t.Run("reject expired token", func(t *testing.T) {
		shortManager := NewManager("test-secret", -1*time.Hour)
		token, _ := shortManager.GenerateToken(1, "test@example.com")

		// Wait a tiny bit to ensure expiration
		time.Sleep(10 * time.Millisecond)

		_, err := manager.ValidateToken(token)
		if err == nil {
			t.Error("expected error for expired token, got nil")
		}
	})
}

func TestJWTManager_TokenExpiration(t *testing.T) {
	t.Run("token has expiration time", func(t *testing.T) {
		expiration := 1 * time.Hour
		manager := NewManager("test-secret", expiration)

		token, err := manager.GenerateToken(1, "test@example.com")
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		claims, err := manager.ValidateToken(token)
		if err != nil {
			t.Fatalf("failed to validate token: %v", err)
		}

		// Check that expiration is set and is in the future
		if claims.ExpiresAt == nil {
			t.Error("expected expiration time to be set")
		}

		expectedExpiry := time.Now().Add(expiration)
		actualExpiry := claims.ExpiresAt.Time

		// Allow 10 second tolerance for test execution time
		timeDiff := actualExpiry.Sub(expectedExpiry)
		if timeDiff < -10*time.Second || timeDiff > 10*time.Second {
			t.Errorf("expiration time mismatch: expected ~%v, got %v (diff: %v)",
				expectedExpiry, actualExpiry, timeDiff)
		}
	})
}

func TestJWTManager_ClaimsContent(t *testing.T) {
	manager := NewManager("test-secret", 24*time.Hour)

	t.Run("claims contain all required fields", func(t *testing.T) {
		userID := uint(456)
		email := "claims@example.com"

		token, err := manager.GenerateToken(userID, email)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		claims, err := manager.ValidateToken(token)
		if err != nil {
			t.Fatalf("failed to validate token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("expected user ID %d, got %d", userID, claims.UserID)
		}
		if claims.Email != email {
			t.Errorf("expected email %s, got %s", email, claims.Email)
		}
		if claims.ExpiresAt == nil {
			t.Error("expected ExpiresAt to be set")
		}
		if claims.IssuedAt == nil {
			t.Error("expected IssuedAt to be set")
		}
		if claims.NotBefore == nil {
			t.Error("expected NotBefore to be set")
		}
	})
}
