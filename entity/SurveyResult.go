package entity

import "time"

type SurveyResult struct {
	Cooldown Cooldown `json:"cooldown"`
	Surveys  []Survey `json:"surveys"`
}

type Survey struct {
	Signature  string    `json:"signature"`
	Symbol     string    `json:"symbol"`
	Deposits   []Deposit `json:"deposits"`
	Expiration time.Time `json:"expiration"`
	Size       string    `json:"size"`
}

type Deposit struct {
	Symbol string `json:"symbol"`
}
