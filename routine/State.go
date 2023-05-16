package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/ui"
	"time"
)

type State struct {
	Agent    *entity.Agent
	Contract *entity.Contract
	Survey   *entity.Survey
	Ship     *entity.Ship

	EventBus chan OrchestratorEvent

	AsleepUntil    *time.Time
	CurrentRoutine Routine
	ForceRoutine   Routine

	WaitingForHttp bool
}

type OrchestratorEvent struct {
	Name string
	Data any
}

func (s *State) Log(message string) {
	go ui.MainLog(fmt.Sprintf("[%s] %s\n", s.Ship.Registration.Name, message))
}

func (s *State) FireEvent(event string, data any) {
	s.EventBus <- OrchestratorEvent{
		Name: event,
		Data: data,
	}
}
