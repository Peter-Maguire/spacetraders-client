package entity

import "time"

type SpaceTraders struct {
	ServerStart   time.Time
	ServerEnd     time.Time
	Orchestrators []Orchestrator
}

type Orchestrator interface {
	GetAgent() *Agent
	GetContract() *Contract
}
