package routine

type Routine interface {
	Run(state *State) RoutineResult
	Name() string
}
