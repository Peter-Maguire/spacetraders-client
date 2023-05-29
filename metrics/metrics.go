package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ContractProgress = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_contract_fulfilled",
		Help: "Items fulfilled in contract",
	})

	ContractRequirement = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_contract_requirement",
		Help: "Items required in contract",
	})
	NumCredits = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "st_agent_credits",
		Help: "Number of credits",
	})
)
