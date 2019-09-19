package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type CrashLoopBackOff_Metrics struct {
	logger     *zap.Logger
	podsCount *prometheus.CounterVec
}

func NewCrashLoopBackOffMetrics(logger *zap.Logger) *CrashLoopBackOff_Metrics {
	return &CrashLoopBackOff_Metrics{logger: logger}

}

func (c *CrashLoopBackOff_Metrics) RegisterMetrics() {
	c.podsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crashloopbackoff_pods_rescheduled",
			Help: "Total number of CrashLoopBackOff Pods rescheduled",
		},
		[]string{"action"},
	)
	prometheus.MustRegister(c.podsCount)
}

func (c *CrashLoopBackOff_Metrics) UpdateRescheduledCount() {
	c.podsCount.With(prometheus.Labels{"action": "rescheduled"}).Inc()
}
