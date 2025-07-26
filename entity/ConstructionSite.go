package entity

type ConstructionSite struct {
	Symbol     Waypoint               `json:"symbol"`
	IsComplete bool                   `json:"isComplete"`
	Materials  []ConstructionMaterial `json:"materials"`
}

func (cs *ConstructionSite) GetMaterial(symbol string) *ConstructionMaterial {
	for _, material := range cs.Materials {
		if material.TradeSymbol == symbol {
			return &material
		}
	}
	return nil
}

type ConstructionMaterial struct {
	TradeSymbol string `json:"tradeSymbol"`
	Required    int    `json:"required"`
	Fulfilled   int    `json:"fulfilled"`
}

func (cm *ConstructionMaterial) IsComplete() bool {
	return cm.Fulfilled >= cm.Required
}
