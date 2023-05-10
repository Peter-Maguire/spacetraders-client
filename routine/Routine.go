package routine

import "spacetraders/entity"

type Routine func(state *entity.State, targetShip *entity.Ship) RoutineResult
