package ui

import "spacetraders/entity"

type Orchestrator interface {
	GetAgent() *entity.Agent
	GetContract() *entity.Contract
}
