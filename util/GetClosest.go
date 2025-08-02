package util

import (
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
)

func GetClosestMarketSelling(selling []string, waypoint entity.LimitedWaypointData) *entity.Waypoint {
	markets := database.GetMarketsSelling(selling)

	if len(markets) == 0 {
		return nil
	}

	// TODO: other system markets
	systemMarkets := make([]database.MarketRates, 0)
	for _, market := range markets {
		if market.Waypoint.GetSystemName() == waypoint.Symbol.GetSystemName() {
			systemMarkets = append(systemMarkets, market)
		}
	}

	sort.Slice(systemMarkets, func(i, j int) bool {
		marketI := markets[i].GetLimitedWaypointData()
		marketJ := markets[j].GetLimitedWaypointData()
		return marketI.GetDistanceFrom(waypoint) < marketJ.GetDistanceFrom(waypoint)
	})

	return &systemMarkets[0].Waypoint
}
