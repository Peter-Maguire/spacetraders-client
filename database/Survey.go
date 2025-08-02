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

func (s *Survey) GetData() entity.Survey {
	data := entity.Survey{}
	json.Unmarshal(s.SurveyData, &data)
	return data
}

func GetUnexpiredSurveysForWaypoint(waypoint entity.Waypoint) []Survey {
	var result []Survey
	db.Where("waypoint = ? AND expires > NOW()", waypoint).Find(&result)
	return result
}
