package repository

import (
	"github.com/roychanmeliaz/btechdevcases/internal/models"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(tx *gorm.DB, transaction *models.Transaction) error
	FindByWalletID(walletID uint, limit int) ([]models.Transaction, error)
	FindByIdempotencyKey(idempotencyKey string) (*models.Transaction, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *gorm.DB, transaction *models.Transaction) error {
	return tx.Create(transaction).Error
}

func (r *transactionRepository) FindByWalletID(walletID uint, limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := r.db.Where("wallet_id = ?", walletID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) FindByIdempotencyKey(idempotencyKey string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Where("idempotency_key = ?", idempotencyKey).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}
