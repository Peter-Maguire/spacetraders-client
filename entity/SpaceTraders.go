package entity

import "time"

type SpaceTraders struct {
	ServerStart   time.Time      `json:"start"`
	ServerEnd     time.Time      `json:"end"`
	Orchestrators []Orchestrator `json:"-"`
}

type Orchestrator interface {
	GetAgent() *Agent
	GetContract() *Contract
}
