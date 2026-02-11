package repository

import (
	"github.com/roychanmeliaz/btechdevcases/internal/models"
	"gorm.io/gorm"
)

type WalletRepository interface {
	Create(wallet *models.Wallet) error
	FindByUserID(userID uint) (*models.Wallet, error)
	UpdateBalance(tx *gorm.DB, walletID uint, newBalance float64) error
	GetBalanceForUpdate(tx *gorm.DB, walletID uint) (float64, error)
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) FindByUserID(userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateBalance(tx *gorm.DB, walletID uint, newBalance float64) error {
	return tx.Model(&models.Wallet{}).
		Where("id = ?", walletID).
		Update("balance", newBalance).Error
}

func (r *walletRepository) GetBalanceForUpdate(tx *gorm.DB, walletID uint) (float64, error) {
	var wallet models.Wallet
	err := tx.Clauses().
		Where("id = ?", walletID).
		First(&wallet).Error
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}
