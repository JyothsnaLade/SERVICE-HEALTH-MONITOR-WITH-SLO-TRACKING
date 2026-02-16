package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	
	"slo-monitor/monitor"
)

var (
	// Health check metrics
	healthCheckTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_health_check_total",
			Help: "Total number of health checks performed",
		},
		[]string{"service", "status"},
	)

	healthCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "slo_health_check_duration_seconds",
			Help:    "Duration of health checks in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service"},
	)

	// SLI metrics
	serviceUptime = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_service_uptime_percent",
			Help: "Current uptime percentage for service",
		},
		[]string{"service"},
	)

	serviceLatencyAvg = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_service_latency_avg_milliseconds",
			Help: "Average latency in milliseconds",
		},
		[]string{"service"},
	)

	serviceLatencyP95 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_service_latency_p95_milliseconds",
			Help: "P95 latency in milliseconds",
		},
		[]string{"service"},
	)

	serviceLatencyP99 = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_service_latency_p99_milliseconds",
			Help: "P99 latency in milliseconds",
		},
		[]string{"service"},
	)

	serviceErrorRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_service_error_rate_percent",
			Help: "Current error rate percentage",
		},
		[]string{"service"},
	)

	serviceTotalRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_service_total_requests",
			Help: "Total requests in evaluation window",
		},
		[]string{"service"},
	)

	// SLO compliance metrics
	sloCompliance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_compliance",
			Help: "SLO compliance status (1 = compliant, 0 = violated)",
		},
		[]string{"service", "slo_name"},
	)

	sloUptimeTarget = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_uptime_target_percent",
			Help: "Configured uptime target",
		},
		[]string{"service", "slo_name"},
	)

	sloLatencyTarget = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_latency_target_milliseconds",
			Help: "Configured latency target",
		},
		[]string{"service", "slo_name"},
	)

	sloErrorBudgetUsed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_error_budget_used_percent",
			Help: "Percentage of error budget consumed",
		},
		[]string{"service", "slo_name"},
	)
)

type MetricsExporter struct{}

func NewMetricsExporter() *MetricsExporter {
	return &MetricsExporter{}
}

func (me *MetricsExporter) RecordHealthCheck(result monitor.CheckResult) {
	service := result.Target.Name
	status := "success"
	if !result.Success {
		status = "failure"
	}

	healthCheckTotal.WithLabelValues(service, status).Inc()
	healthCheckDuration.WithLabelValues(service).Observe(result.Latency.Seconds())
}

func (me *MetricsExporter) UpdateSLI(sli monitor.SLI) {
	service := sli.ServiceName

	serviceUptime.WithLabelValues(service).Set(sli.Uptime)
	serviceLatencyAvg.WithLabelValues(service).Set(sli.AvgLatency)
	serviceLatencyP95.WithLabelValues(service).Set(sli.P95Latency)
	serviceLatencyP99.WithLabelValues(service).Set(sli.P99Latency)
	serviceErrorRate.WithLabelValues(service).Set(sli.ErrorRate)
	serviceTotalRequests.WithLabelValues(service).Set(float64(sli.TotalRequests))
}

func (me *MetricsExporter) UpdateSLOStatus(status monitor.SLOStatus) {
	service := status.SLI.ServiceName
	sloName := status.SLO.Name

	compliance := 0.0
	if status.Compliant {
		compliance = 1.0
	}

	sloCompliance.WithLabelValues(service, sloName).Set(compliance)
	sloUptimeTarget.WithLabelValues(service, sloName).Set(status.SLO.UptimeTarget)
	sloLatencyTarget.WithLabelValues(service, sloName).Set(float64(status.SLO.LatencyTarget))
	sloErrorBudgetUsed.WithLabelValues(service, sloName).Set(status.ErrorBudgetUsed)
}