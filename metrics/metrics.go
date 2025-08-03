package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ContractProgress = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_contract_fulfilled",
		Help: "Items fulfilled in contract",
	}, []string{"agent"})

	ContractRequirement = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_contract_requirement",
		Help: "Items required in contract",
	}, []string{"agent"})
	NumCredits = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "st_agent_credits",
		Help: "Number of credits",
	}, []string{"agent"})
)
