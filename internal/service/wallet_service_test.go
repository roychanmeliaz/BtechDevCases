package service

import (
	"errors"
	"testing"

	"github.com/roychanmeliaz/btechdevcases/internal/models"
	"github.com/roychanmeliaz/btechdevcases/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupWalletTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Wallet{}, &models.Transaction{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func createTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	user := &models.User{
		Email:    email,
		Password: "hashedpassword",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	wallet := &models.Wallet{
		UserID:  user.ID,
		Balance: 1000.0,
	}
	if err := db.Create(wallet).Error; err != nil {
		t.Fatalf("failed to create test wallet: %v", err)
	}

	return user
}

func TestWalletService_GetWallet(t *testing.T) {
	db := setupWalletTestDB(t)
	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	walletService := NewWalletService(userRepo, walletRepo, transactionRepo, db)

	user := createTestUser(t, db, "wallet@example.com")

	t.Run("get wallet with no transactions", func(t *testing.T) {
		wallet, transactions, err := walletService.GetWallet(user.ID)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if wallet == nil {
			t.Error("expected wallet, got nil")
		}
		if wallet.Balance != 1000.0 {
			t.Errorf("expected balance 1000, got %f", wallet.Balance)
		}
		if len(transactions) != 0 {
			t.Errorf("expected 0 transactions, got %d", len(transactions))
		}
	})
}

func TestWalletService_Transfer(t *testing.T) {
	db := setupWalletTestDB(t)
	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	walletService := NewWalletService(userRepo, walletRepo, transactionRepo, db)

	sender := createTestUser(t, db, "sender@example.com")
	recipient := createTestUser(t, db, "recipient@example.com")

	t.Run("successful transfer", func(t *testing.T) {
		err := walletService.Transfer(sender.ID, recipient.Email, 200.0, "Test payment", "test-key-1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Verify sender balance
		senderWallet, err := walletRepo.FindByUserID(sender.ID)
		if err != nil {
			t.Errorf("failed to get sender wallet: %v", err)
		}
		if senderWallet.Balance != 800.0 {
			t.Errorf("expected sender balance 800, got %f", senderWallet.Balance)
		}

		// Verify recipient balance
		recipientWallet, err := walletRepo.FindByUserID(recipient.ID)
		if err != nil {
			t.Errorf("failed to get recipient wallet: %v", err)
		}
		if recipientWallet.Balance != 1200.0 {
			t.Errorf("expected recipient balance 1200, got %f", recipientWallet.Balance)
		}

		// Verify transactions were created
		senderTxs, err := transactionRepo.FindByWalletID(senderWallet.ID, 10)
		if err != nil {
			t.Errorf("failed to get sender transactions: %v", err)
		}
		if len(senderTxs) != 1 {
			t.Errorf("expected 1 sender transaction, got %d", len(senderTxs))
		}
		if senderTxs[0].Type != models.TransactionTypeDebit {
			t.Errorf("expected debit transaction, got %s", senderTxs[0].Type)
		}

		recipientTxs, err := transactionRepo.FindByWalletID(recipientWallet.ID, 10)
		if err != nil {
			t.Errorf("failed to get recipient transactions: %v", err)
		}
		if len(recipientTxs) != 1 {
			t.Errorf("expected 1 recipient transaction, got %d", len(recipientTxs))
		}
		if recipientTxs[0].Type != models.TransactionTypeCredit {
			t.Errorf("expected credit transaction, got %s", recipientTxs[0].Type)
		}
	})

	t.Run("insufficient balance", func(t *testing.T) {
		err := walletService.Transfer(sender.ID, recipient.Email, 10000.0, "Too much", "test-key-2")
		if !errors.Is(err, ErrInsufficientBalance) {
			t.Errorf("expected ErrInsufficientBalance, got %v", err)
		}
	})

	t.Run("recipient not found", func(t *testing.T) {
		err := walletService.Transfer(sender.ID, "nonexistent@example.com", 50.0, "To nobody", "test-key-3")
		if !errors.Is(err, ErrRecipientNotFound) {
			t.Errorf("expected ErrRecipientNotFound, got %v", err)
		}
	})

	t.Run("self transfer", func(t *testing.T) {
		err := walletService.Transfer(sender.ID, sender.Email, 50.0, "To myself", "test-key-4")
		if !errors.Is(err, ErrSelfTransfer) {
			t.Errorf("expected ErrSelfTransfer, got %v", err)
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		err := walletService.Transfer(sender.ID, recipient.Email, -50.0, "Negative", "test-key-5")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}

		err = walletService.Transfer(sender.ID, recipient.Email, 0, "Zero", "test-key-6")
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("idempotency", func(t *testing.T) {
		// Get current balance
		senderWallet, _ := walletRepo.FindByUserID(sender.ID)
		initialBalance := senderWallet.Balance

		// First transfer
		err := walletService.Transfer(sender.ID, recipient.Email, 50.0, "First", "idempotent-key")
		if err != nil {
			t.Errorf("expected no error on first transfer, got %v", err)
		}

		// Verify balance changed
		senderWallet, _ = walletRepo.FindByUserID(sender.ID)
		afterFirstBalance := senderWallet.Balance
		if afterFirstBalance != initialBalance-50.0 {
			t.Errorf("expected balance %f after first transfer, got %f", initialBalance-50.0, afterFirstBalance)
		}

		// Duplicate transfer with same idempotency key
		err = walletService.Transfer(sender.ID, recipient.Email, 50.0, "Duplicate", "idempotent-key")
		if err != nil {
			t.Errorf("expected no error on duplicate transfer, got %v", err)
		}

		// Verify balance did not change
		senderWallet, _ = walletRepo.FindByUserID(sender.ID)
		finalBalance := senderWallet.Balance
		if finalBalance != afterFirstBalance {
			t.Errorf("expected balance %f after duplicate transfer, got %f", afterFirstBalance, finalBalance)
		}
	})
}
