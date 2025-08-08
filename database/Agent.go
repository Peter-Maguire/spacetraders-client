package database

import "time"

type Agent struct {
	Symbol    string `gorm:"primary_key"`
	Enabled   bool   `gorm:"default:true"`
	Token     string
	Config    map[string]any `gorm:"type:json"`
	CreatedAt time.Time
}
