package database

import (
	"encoding/json"
	"spacetraders/entity"
	"time"
)

type Survey struct {
	Waypoint   string
	SurveyData json.RawMessage
	Date       time.Time
	Expires    time.Time
}

func (s *Survey) StoreSurvey(waypoint string, survey entity.Survey) {

}
