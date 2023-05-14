package database

import (
	"encoding/json"
	"spacetraders/entity"
	"time"
)

type Survey struct {
	Waypoint   string
	SurveyData []byte `gorm:"type:json"`
	Date       time.Time
	Expires    time.Time
}

func StoreSurvey(waypoint entity.Waypoint, survey entity.Survey) {
	surveyData, _ := json.Marshal(survey)
	db.Create(Survey{
		Waypoint:   string(waypoint),
		SurveyData: surveyData,
		Date:       time.Now(),
		Expires:    survey.Expiration,
	})
}
