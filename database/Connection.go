package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var db *gorm.DB

func Init() {
	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{})
	if err != nil {
		fmt.Println("db error", err)
		os.Exit(1)
	}

	err = db.AutoMigrate(&MarketRates{}, &ShipCost{})
	if err != nil {
		fmt.Println("automigrate error", err)
		os.Exit(1)
	}
}
