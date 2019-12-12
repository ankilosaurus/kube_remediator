package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type CrashLoopBackOff_Metrics struct {
	logger     *zap.Logger
	pods_count *prometheus.CounterVec
}

func NewCrashLoopBackOffMetrics(logger *zap.Logger) *CrashLoopBackOff_Metrics {
	return &CrashLoopBackOff_Metrics{logger: logger}

}

func (c *CrashLoopBackOff_Metrics) Register() {
	c.pods_count = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crashloopbackoff_pods_rescheduled",
			Help: "Total number of CrashLoopBackOff Pods",
		},
		[]string{"action"},
	)
	prometheus.MustRegister(c.pods_count)
}

func (c *CrashLoopBackOff_Metrics) UnRegister() {
	prometheus.Unregister(c.pods_count)
}

func (c *CrashLoopBackOff_Metrics) UpdateRescheduledCount() {
	c.pods_count.With(prometheus.Labels{"action": "rescheduled"}).Inc()
}
