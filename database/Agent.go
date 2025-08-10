package database

import "time"

type Agent struct {
	Symbol    string `gorm:"primary_key"`
	Enabled   bool   `gorm:"default:true"`
	Token     string
	Config    AgentConfig `gorm:"type:json"`
	CreatedAt time.Time
}

type AgentConfig map[string]any

func (ac AgentConfig) GetString(key string, def string) string {
	val, ok := ac[key]
	if !ok {
		return def
	}
	return val.(string)
}

func (ac AgentConfig) GetInt(key string, def int) int {
	val, ok := ac[key]
	if !ok {
		return def
	}
	return val.(int)
}

func (ac AgentConfig) GetBool(key string, def bool) bool {
	val, ok := ac[key]
	if !ok {
		return def
	}
	return val.(bool)
}

func GetEnabledAgents() []Agent {
	var agents []Agent
	db.Find(&agents, "enabled = true")
	return agents
}

func SetAgentToken(symbol string, token string) {
	db.Model(&Agent{}).Where("symbol = ?", symbol).Update("token", token)
}
