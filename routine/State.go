package routine

import (
	"context"
	"fmt"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/ui"
	"sync"
	"time"
)

type Phaser interface {
	GetPhase() constant.Phase
}

type State struct {
	Agent            *entity.Agent            `json:"-"`
	Contract         *entity.Contract         `json:"-"`
	ConstructionSite *entity.ConstructionSite `json:"-"`
	Survey           *entity.Survey
	Ship             *entity.Ship
	Config           *database.AgentConfig

	Haulers  []*entity.Ship
	EventBus *chan OrchestratorEvent

	AsleepUntil     *time.Time
	WaitingForEvent string
	CurrentRoutine  Routine
	ForceRoutine    Routine

	States *[]*State

	Phase Phaser

	StatesMutex *sync.Mutex

	WaitingForHttp bool
	StoppedReason  string
	Context        context.Context
}

func (s *State) SetWaitingForHttp(b bool) {
	s.WaitingForHttp = b
}

type OrchestratorEvent struct {
	Name string
	Data any
}

func (s *State) Log(message string) {
	ui.MainLog(fmt.Sprintf("[%s] %s", s.Ship.Registration.Name, message))
}

func (s *State) FireEvent(event string, data any) {
	*s.EventBus <- OrchestratorEvent{
		Name: event,
		Data: data,
	}
}

func (s *State) GetShipsWithRole(t constant.ShipRole) []*entity.Ship {
	ships := make([]*entity.Ship, 0)
	for _, s := range *s.States {
		if s.Ship.Registration.Role == t {
			ships = append(ships, s.Ship)
		}
	}
	return ships
}

func (s *State) GetShipsWithRoleAtOrGoingToWaypoint(t constant.ShipRole, waypoint entity.Waypoint) []*entity.Ship {
	ships := make([]*entity.Ship, 0)
	for _, s := range *s.States {
		if s.Ship.Registration.Role == t && (s.Ship.Nav.WaypointSymbol == waypoint || (s.Ship.Nav.Status == "IN_TRANSIT" && s.Ship.Nav.Route.Destination.Symbol == waypoint)) {
			ships = append(ships, s.Ship)
		}
	}
	return ships
}

func (s *State) GetTotalOfItemAcrossAllShips(symbol string) int {
	amount := 0
	for _, state := range *s.States {
		slot := state.Ship.Cargo.GetSlotWithItem(symbol)
		if slot == nil {
			continue
		}
		amount += slot.Units
	}
	return amount
}

func (s *State) GetAgent() *entity.Agent {
	return s.Agent
}

func (s *State) GetShip() *entity.Ship {
	return s.Ship
}
