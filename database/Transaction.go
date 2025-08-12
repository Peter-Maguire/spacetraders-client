package database

import (
	"spacetraders/entity"
)

type Transaction struct {
	ID     int64 `gorm:"primary_key;AUTO_INCREMENT"`
	Source string
	entity.MarketTransaction
}

func LogTransaction(source string, tx entity.MarketTransaction) {
	db.Create(&Transaction{
		Source:            source,
		MarketTransaction: tx,
	})
}
