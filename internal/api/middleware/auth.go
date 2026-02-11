package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	customjwt "github.com/roychanmeliaz/btechdevcases/pkg/jwt"
)

type AuthMiddleware struct {
	jwtManager     *customjwt.Manager
	redisClient    *redis.Client
	sessionTimeout time.Duration
}

func NewAuthMiddleware(jwtManager *customjwt.Manager, redisClient *redis.Client, sessionTimeout time.Duration) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:     jwtManager,
		redisClient:    redisClient,
		sessionTimeout: sessionTimeout,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate JWT token
		claims, err := m.jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Check session in Redis
		ctx := context.Background()
		sessionKey := "session:" + token
		userID, err := m.redisClient.Get(ctx, sessionKey).Result()
		if err == redis.Nil {
			// Session expired or doesn't exist
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired due to inactivity"})
			c.Abort()
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error checking session"})
			c.Abort()
			return
		}

		// Update session activity (reset expiration timer)
		err = m.redisClient.Set(ctx, sessionKey, userID, m.sessionTimeout).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating session"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
