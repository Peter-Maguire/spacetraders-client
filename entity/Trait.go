package entity

import "spacetraders/constant"

type Trait struct {
	Symbol      constant.WaypointTrait `json:"symbol"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
}
