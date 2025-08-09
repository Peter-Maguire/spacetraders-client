package entity

import (
	"time"
)

type Market struct {
	Symbol       Waypoint            `json:"symbol"`
	Imports      []ExchangeItem      `json:"imports"`
	Exports      []ExchangeItem      `json:"exports"`
	Exchange     []ExchangeItem      `json:"exchange"`
	Transactions []MarketTransaction `json:"transactions"`
	TradeGoods   []MarketGood        `json:"tradeGoods"`
}

func (m *Market) GetTradeGood(itemSymbol string) *MarketGood {
	for _, tradeGood := range m.TradeGoods {
		if tradeGood.Symbol == itemSymbol {
			return &tradeGood
		}
	}
	return nil
}

type MarketTransaction struct {
	WaypointSymbol Waypoint  `json:"waypointSymbol"`
	ShipSymbol     string    `json:"shipSymbol"`
	TradeSymbol    string    `json:"tradeSymbol"`
	Type           string    `json:"type"`
	Units          int       `json:"units"`
	Price          int       `json:"price"`
	PricePerUnit   int       `json:"pricePerUnit"`
	TotalPrice     int       `json:"totalPrice"`
	Timestamp      time.Time `json:"timestamp"`
}

type MarketGood struct {
	Symbol        string `json:"symbol"`
	TradeVolume   int    `json:"tradeVolume"`
	Supply        string `json:"supply"`
	PurchasePrice int    `json:"purchasePrice"`
	SellPrice     int    `json:"sellPrice"`
}

type ExchangeItem struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
