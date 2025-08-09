package database

import (
	"spacetraders/entity"
)

type Transaction struct {
	ID int64 `gorm:"primary_key;AUTO_INCREMENT"`
	entity.MarketTransaction
}

func LogTransaction(tx entity.MarketTransaction) {
	db.Create(&Transaction{
		MarketTransaction: tx,
	})
}
