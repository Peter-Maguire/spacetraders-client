package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var db *gorm.DB

func Init() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,         // Don't include params in the SQL log
			Colorful:                  false,        // Disable color
		},
	)

	var err error
	db, err = gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		fmt.Println("db error", err)
		os.Exit(1)
	}

	err = db.AutoMigrate(&MarketRates{}, &ShipCost{}, &MarketExchange{} /*&Waypoint{}, &System{}, &Survey{}*/)
	if err != nil {

		fmt.Println("automigrate error", err)
		os.Exit(1)
	}
}

func Reset() {
	fmt.Println("Resetting database")
	db.Where("1 = 1").Delete(&MarketRates{})
	db.Where("1 = 1").Delete(&MarketExchange{})
	db.Where("1 = 1").Delete(&ShipCost{})
	db.Where("1 = 1").Delete(&Survey{})
	db.Where("1 = 1").Delete(&System{})
	db.Where("1 = 1").Delete(&Waypoint{})
}
