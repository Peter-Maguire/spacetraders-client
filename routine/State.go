package routine

import (
	"context"
	"fmt"
	"spacetraders/constant"
	"spacetraders/entity"
	"spacetraders/ui"
	"sync"
	"time"
)

type State struct {
	Agent    *entity.Agent    `json:"-"`
	Contract *entity.Contract `json:"-"`
	Survey   *entity.Survey
	Ship     *entity.Ship

	Haulers  []*entity.Ship
	EventBus chan OrchestratorEvent

	AsleepUntil    *time.Time
	CurrentRoutine Routine
	ForceRoutine   Routine

	States *[]*State

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
	s.EventBus <- OrchestratorEvent{
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
