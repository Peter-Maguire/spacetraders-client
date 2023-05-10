package entity

type ContractDeliverResult struct {
	Contract Contract  `json:"contract"`
	Cargo    ShipCargo `json:"cargo"`
}
