package routine

import "time"

type RoutineResult struct {
	SetRoutine  Routine
	WaitSeconds int
	WaitUntil   *time.Time
	Stop        bool
}
