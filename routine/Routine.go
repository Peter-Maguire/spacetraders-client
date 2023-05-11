package routine

import "spacetraders/entity"

type Routine func(state *entity.State) RoutineResult
