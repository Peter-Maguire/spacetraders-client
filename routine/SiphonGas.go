package routine

import (
	"fmt"
	"spacetraders/constant"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
)

type SiphonGas struct {
	latentCooldown *time.Time
}

func (f SiphonGas) Run(state *State) RoutineResult {

	if f.latentCooldown != nil && f.latentCooldown.After(time.Now()) {
		state.Log("Waiting for Cooldown")
		return RoutineResult{WaitUntil: f.latentCooldown}
	}
	waypoint := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
	waypointData := waypoint.GetData()

	if waypointData.Type != constant.WaypointTypeGasGiant {
		state.Log("Not at a gas giant")
		return RoutineResult{
			SetRoutine: GoToGasGiant{},
		}
	}

	result, err := state.Ship.Siphon(state.Context)
	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			state.Log("We are on cooldown from a previous running routine")
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		case http.ErrCargoFull:
			return RoutineResult{
				SetRoutine: Jettison{
					nextIfFailed:     FullWait{},
					nextIfSuccessful: f,
				},
			}
		case http.ErrCannotExtractHere:
			state.Log("We're not at an asteroid field")
			return RoutineResult{
				SetRoutine: GoToGasGiant{},
			}
		case http.ErrShipSurveyExhausted, http.ErrShipSurveyVerification, http.ErrShipSurveyExpired:
			state.Log("Something went wrong with the survey " + err.Error())
			state.FireEvent("surveyExhausted", state.Survey)
			state.Survey = nil
			return RoutineResult{}
		case http.ErrOverExtracted:
			state.Log("Asteroid Over-extracted")
			return RoutineResult{
				SetRoutine: GoToGasGiant{blacklist: []entity.Waypoint{state.Ship.Nav.WaypointSymbol}},
			}
		}

		state.Log(fmt.Sprintf("Unknown error: %s", err))
		// No idea
		return RoutineResult{
			WaitSeconds: 10,
		}
	}
	for _, event := range result.Events {
		state.Log(fmt.Sprintf("!!! Siphon Event - %s: %s", event.Name, event.Description))
	}

	if state.Ship.Cargo.IsFull() {
		return RoutineResult{
			SetRoutine: Jettison{
				nextIfFailed: FullWait{},
				nextIfSuccessful: SiphonGas{
					latentCooldown: &result.Cooldown.Expiration,
				},
			},
		}
	}

	return RoutineResult{
		WaitUntil: &result.Cooldown.Expiration,
	}

}

func (f SiphonGas) Name() string {
	return "Siphon Gas"
}
