package entity

import (
	"fmt"
	"spacetraders/http"
	"strings"
)

type Waypoint string

func (w *Waypoint) GetSystem() string {
	strw := string(*w)
	return strw[:strings.LastIndex(strw, "-")]
}

func (w *Waypoint) GetMarket() (*Market, error) {
	return http.Request[Market]("GET", fmt.Sprintf("systems/%s/waypoints/%s/market", w.GetSystem(), *w), nil)
}
