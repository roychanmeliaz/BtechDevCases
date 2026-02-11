package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/roychanmeliaz/btechdevcases/internal/api/handlers"
	"github.com/roychanmeliaz/btechdevcases/internal/api/middleware"
	"github.com/roychanmeliaz/btechdevcases/internal/service"
	customjwt "github.com/roychanmeliaz/btechdevcases/pkg/jwt"
)

type Router struct {
	engine          *gin.Engine
	authHandler     *handlers.AuthHandler
	walletHandler   *handlers.WalletHandler
	authMiddleware  *middleware.AuthMiddleware
}

func NewRouter(
	authService service.AuthService,
	walletService service.WalletService,
	jwtManager *customjwt.Manager,
	redisClient *redis.Client,
	sessionTimeout time.Duration,
) *Router {
	// Create handlers
	authHandler := handlers.NewAuthHandler(authService, redisClient, sessionTimeout)
	walletHandler := handlers.NewWalletHandler(walletService)
	
	// Create middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, redisClient, sessionTimeout)

	// Setup Gin engine
	engine := gin.Default()

	return &Router{
		engine:          engine,
		authHandler:     authHandler,
		walletHandler:   walletHandler,
		authMiddleware:  authMiddleware,
	}
}

func (r *Router) Setup() {
	// Health check endpoint
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.engine.Group("/api")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(r.authMiddleware.RequireAuth())
		{
			// User endpoints
			protected.GET("/me", r.walletHandler.GetMe)

			// Wallet endpoints
			protected.GET("/wallet", r.walletHandler.GetWallet)
			protected.POST("/wallet/transfer", r.walletHandler.Transfer)
		}
	}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
