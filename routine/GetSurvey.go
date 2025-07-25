package routine

import (
	"fmt"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/ui"
)

type GetSurvey struct {
	next Routine
}

func (g GetSurvey) Run(state *State) RoutineResult {
	if g.next == nil {
		g.next = MineOres{}
	}

	if !state.Ship.HasMount("MOUNT_SURVEYOR_I") || state.Survey != nil || state.Contract == nil {
		state.Log("No surveyor mount or survey exists")
		return RoutineResult{
			SetRoutine: g.next,
		}
	}

	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)
	state.Log("Finding a survey")
	surveyResult, err := state.Ship.Survey(state.Context)

	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			state.Log("We are on cooldown from a previous running routine")
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		}
		state.Log(fmt.Sprintf("Unknown error: %s", err))
		// No idea
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	//fmt.Println(surveyResult)

	for _, survey := range surveyResult.Surveys {
		database.StoreSurvey(state.Ship.Nav.WaypointSymbol, survey)
	}

	if state.Contract == nil {
		state.Survey = &surveyResult.Surveys[0]
		state.FireEvent("goodSurveyFound", state.Survey)
	} else {
		bestSurvey := findBestSurvey(surveyResult.Surveys, state.Contract.Terms.Deliver)

		if bestSurvey != nil {
			state.Survey = bestSurvey
			state.Log(fmt.Sprintf("Good survey found: %s\n", bestSurvey.Signature))
			state.FireEvent("goodSurveyFound", state.Survey)
		} else {
			state.Log("No survey available that satisfies our needs")
		}
	}

	ui.MainLog(fmt.Sprintf("Waiting %d seconds", surveyResult.Cooldown.RemainingSeconds))

	return RoutineResult{
		SetRoutine: g.next,
		WaitUntil:  &surveyResult.Cooldown.Expiration,
	}
}

func (g GetSurvey) Name() string {
	return "Get Survey"
}

func findBestSurvey(surveys []entity.Survey, deliverables []entity.ContractDeliverable) *entity.Survey {
	for _, survey := range surveys {
		for _, deposit := range survey.Deposits {
			//fmt.Printf("Survey %s has deposit of %s\n", survey.Signature, deposit.Symbol)
			for _, deliverable := range deliverables {
				if deposit.Symbol == deliverable.TradeSymbol {
					return &survey
				}
			}
		}
	}
	return nil
}
