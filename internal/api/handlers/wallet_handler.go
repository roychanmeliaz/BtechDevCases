package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/roychanmeliaz/btechdevcases/internal/service"
)

type WalletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

type TransferRequest struct {
	Recipient string  `json:"recipient" binding:"required,email"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Notes     string  `json:"notes"`
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	wallet, transactions, err := h.walletService.GetWallet(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching wallet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet":       wallet,
		"transactions": transactions,
	})
}

func (h *WalletHandler) Transfer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate or get idempotency key from header
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		// Generate one if not provided (for backwards compatibility)
		idempotencyKey = uuid.New().String()
	}

	err := h.walletService.Transfer(userID.(uint), req.Recipient, req.Amount, req.Notes, idempotencyKey)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientBalance):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrRecipientNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidAmount):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrSelfTransfer):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing transfer"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "transfer successful",
	})
}

func (h *WalletHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	email, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Hello " + email.(string) + ", welcome back",
		"user_id": userID,
	})
}
