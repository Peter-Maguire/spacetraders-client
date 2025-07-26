package util

import (
	"math"
	"sort"
	"spacetraders/entity"
)

func CalcDistance(x1 int, y1 int, x2 int, y2 int) int {
	return int(math.Sqrt(math.Pow(float64(x1-x2), 2) + math.Pow(float64(y1-y2), 2)))
}

func SortWaypointsClosestTo(waypoints []*entity.WaypointData, waypoint entity.LimitedWaypointData) {
	sort.Slice(waypoints, func(i, j int) bool {
		d1 := waypoints[i].GetDistanceFrom(waypoint)
		d2 := waypoints[j].GetDistanceFrom(waypoint)
		return d1 < d2
	})
}
