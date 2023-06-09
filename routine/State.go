package routine

import (
	"fmt"
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
