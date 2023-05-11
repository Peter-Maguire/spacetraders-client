package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
)

func GetSurvey(state *entity.State) RoutineResult {
	state.Survey = nil
	_ = state.Ship.EnsureNavState(entity.NavOrbit)
	fmt.Println("Finding a survey")
	surveyResult, err := state.Ship.Survey()

	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			fmt.Println("We are on cooldown from a previous running routine")
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		}
		fmt.Println("Unknown error", err.Data)
		// No idea
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	fmt.Println(surveyResult)

	bestSurvey := findBestSurvey(surveyResult.Surveys, state.Contract.Terms.Deliver)

	if bestSurvey != nil {
		state.Survey = bestSurvey
		fmt.Printf("Good survey found: %s\n", bestSurvey.Signature)
	} else {
		fmt.Println("No survey available that satisfies our needs")
	}

	fmt.Printf("Waiting %d seconds\n", surveyResult.Cooldown.RemainingSeconds)

	return RoutineResult{
		SetRoutine:  MineOres,
		WaitSeconds: surveyResult.Cooldown.RemainingSeconds,
	}
}

func findBestSurvey(surveys []entity.Survey, deliverables []entity.ContractDeliverable) *entity.Survey {
	for _, survey := range surveys {
		for _, deposit := range survey.Deposits {
			fmt.Printf("Survey %s has deposit of %s\n", survey.Signature, deposit.Symbol)
			for _, deliverable := range deliverables {
				if deposit.Symbol == deliverable.TradeSymbol {
					return &survey
				}
			}
		}
	}
	return nil
}
