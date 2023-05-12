package routine

import (
	"fmt"
	"spacetraders/entity"
)

type State struct {
	Agent    *entity.Agent
	Contract *entity.Contract
	Survey   *entity.Survey
	Ship     *entity.Ship

	EventBus chan OrchestratorEvent

	ForceRoutine         Routine
	LastVisitedWaypoints []entity.Waypoint
}

type OrchestratorEvent struct {
	Name string
	Data any
}

func (s *State) Log(message string) {
	fmt.Printf("[%s] %s\n", s.Ship.Registration.Name, message)
}

func (s *State) FireEvent(event string, data any) {
	s.EventBus <- OrchestratorEvent{
		Name: event,
		Data: data,
	}
}
