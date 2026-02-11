package service

import (
	"errors"
	"fmt"

	"github.com/roychanmeliaz/btechdevcases/internal/models"
	"github.com/roychanmeliaz/btechdevcases/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrRecipientNotFound   = errors.New("recipient not found")
	ErrInvalidAmount       = errors.New("amount must be greater than 0")
	ErrSelfTransfer        = errors.New("cannot transfer to yourself")
)

type WalletService interface {
	GetWallet(userID uint) (*models.Wallet, []models.Transaction, error)
	Transfer(senderID uint, recipientEmail string, amount float64, notes string, idempotencyKey string) error
}

type walletService struct {
	userRepo        repository.UserRepository
	walletRepo      repository.WalletRepository
	transactionRepo repository.TransactionRepository
	db              *gorm.DB
}

func NewWalletService(
	userRepo repository.UserRepository,
	walletRepo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	db *gorm.DB,
) WalletService {
	return &walletService{
		userRepo:        userRepo,
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

func (s *walletService) GetWallet(userID uint) (*models.Wallet, []models.Transaction, error) {
	// Get wallet
	wallet, err := s.walletRepo.FindByUserID(userID)
	if err != nil {
		return nil, nil, fmt.Errorf("error finding wallet: %w", err)
	}

	// Get recent transactions (limit 50)
	transactions, err := s.transactionRepo.FindByWalletID(wallet.ID, 50)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, fmt.Errorf("error finding transactions: %w", err)
	}

	return wallet, transactions, nil
}

func (s *walletService) Transfer(senderID uint, recipientEmail string, amount float64, notes string, idempotencyKey string) error {
	// Check for duplicate request using idempotency key
	if idempotencyKey != "" {
		existingTx, err := s.transactionRepo.FindByIdempotencyKey(idempotencyKey)
		if err == nil && existingTx != nil {
			// Transaction already processed, return success (idempotent)
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("error checking idempotency: %w", err)
		}
	}

	// Validate amount
	if amount <= 0 {
		return ErrInvalidAmount
	}

	// Find sender
	sender, err := s.userRepo.FindByID(senderID)
	if err != nil {
		return fmt.Errorf("error finding sender: %w", err)
	}

	// Find recipient
	recipient, err := s.userRepo.FindByEmail(recipientEmail)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecipientNotFound
		}
		return fmt.Errorf("error finding recipient: %w", err)
	}

	// Check self-transfer
	if sender.ID == recipient.ID {
		return ErrSelfTransfer
	}

	// Execute transfer in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get sender wallet with row lock
		senderWallet, err := s.walletRepo.FindByUserID(sender.ID)
		if err != nil {
			return fmt.Errorf("error finding sender wallet: %w", err)
		}

		// Get recipient wallet with row lock
		recipientWallet, err := s.walletRepo.FindByUserID(recipient.ID)
		if err != nil {
			return fmt.Errorf("error finding recipient wallet: %w", err)
		}

		// Check sufficient balance
		if senderWallet.Balance < amount {
			return ErrInsufficientBalance
		}

		// Update balances
		newSenderBalance := senderWallet.Balance - amount
		newRecipientBalance := recipientWallet.Balance + amount

		if err := s.walletRepo.UpdateBalance(tx, senderWallet.ID, newSenderBalance); err != nil {
			return fmt.Errorf("error updating sender balance: %w", err)
		}

		if err := s.walletRepo.UpdateBalance(tx, recipientWallet.ID, newRecipientBalance); err != nil {
			return fmt.Errorf("error updating recipient balance: %w", err)
		}

		// Create debit transaction for sender
		debitTx := &models.Transaction{
			WalletID:       senderWallet.ID,
			Amount:         amount,
			Type:           models.TransactionTypeDebit,
			RelatedUserID:  &recipient.ID,
			Notes:          notes,
			IdempotencyKey: idempotencyKey,
		}
		if err := s.transactionRepo.Create(tx, debitTx); err != nil {
			return fmt.Errorf("error creating debit transaction: %w", err)
		}

		// Create credit transaction for recipient
		creditTx := &models.Transaction{
			WalletID:       recipientWallet.ID,
			Amount:         amount,
			Type:           models.TransactionTypeCredit,
			RelatedUserID:  &sender.ID,
			Notes:          notes,
			IdempotencyKey: idempotencyKey + "-credit", // Different key for credit side
		}
		if err := s.transactionRepo.Create(tx, creditTx); err != nil {
			return fmt.Errorf("error creating credit transaction: %w", err)
		}

		return nil
	})
}
