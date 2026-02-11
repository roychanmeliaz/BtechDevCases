package models

import (
	"time"

	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeDebit  TransactionType = "debit"
	TransactionTypeCredit TransactionType = "credit"
)

type Transaction struct {
	ID             uint            `gorm:"primarykey" json:"id"`
	WalletID       uint            `gorm:"not null;index" json:"wallet_id"`
	Amount         float64         `gorm:"not null" json:"amount"`
	Type           TransactionType `gorm:"not null;type:varchar(10)" json:"type"`
	RelatedUserID  *uint           `gorm:"index" json:"related_user_id,omitempty"`
	Notes          string          `gorm:"type:text" json:"notes,omitempty"`
	IdempotencyKey string          `gorm:"uniqueIndex;type:varchar(255)" json:"-"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
	Wallet         *Wallet         `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
	RelatedUser    *User           `gorm:"foreignKey:RelatedUserID" json:"related_user,omitempty"`
}

func (Transaction) TableName() string {
	return "transactions"
}
