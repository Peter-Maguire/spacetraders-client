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

	LastVisitedWaypoints []entity.Waypoint
}

func (s *State) Log(message string) {
	fmt.Printf("[%s] %s\n", s.Ship.Registration.Name, message)
}
