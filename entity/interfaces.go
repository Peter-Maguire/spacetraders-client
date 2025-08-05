package entity

type state interface {
	GetAgent() *Agent
	GetShip() *Ship
}
