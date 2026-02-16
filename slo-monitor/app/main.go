package main

import (
	"encoding/json"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"

	"slo-monitor/alerts"
	"slo-monitor/config"
	"slo-monitor/metrics"
	"slo-monitor/monitor"
)

func main() {
	log.Println("🚀 Starting SLO Monitor...")

	// Load configuration
	cfg := config.Load()
	log.Printf("✓ Configuration loaded - Check interval: %s", cfg.CheckInterval)

	// Load targets
	targets, err := loadTargets(cfg.TargetsFile)
	if err != nil {
		log.Fatalf("Failed to load targets: %v", err)
	}
	log.Printf("✓ Loaded %d targets", len(targets))

	// Load SLOs
	slos, err := loadSLOs(cfg.SLOsFile)
	if err != nil {
		log.Fatalf("Failed to load SLOs: %v", err)
	}
	log.Printf("✓ Loaded %d SLOs", len(slos))

	// Initialize components
	checker := monitor.NewHealthChecker()
	sliCalculator := monitor.NewSLICalculator(24 * time.Hour) // 24 hour window
	sloEvaluator := monitor.NewSLOEvaluator(slos)
	metricsExporter := metrics.NewMetricsExporter()
	alerter := alerts.NewAlerter(cfg.AlertWebhookURL)

	// Start Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/status", statusHandler(sliCalculator, sloEvaluator))

	go func() {
		log.Printf("✓ Metrics server starting on :%s", cfg.PrometheusPort)
		if err := http.ListenAndServe(":"+cfg.PrometheusPort, nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Start monitoring loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("✓ Monitoring started")

	// Run first check immediately
	runMonitoringCycle(ctx, checker, sliCalculator, sloEvaluator, metricsExporter, alerter, targets)

	for {
		select {
		case <-ticker.C:
			runMonitoringCycle(ctx, checker, sliCalculator, sloEvaluator, metricsExporter, alerter, targets)
		case <-sigChan:
			log.Println("Shutting down gracefully...")
			return
		}
	}
}

func runMonitoringCycle(
	ctx context.Context,
	checker *monitor.HealthChecker,
	sliCalculator *monitor.SLICalculator,
	sloEvaluator *monitor.SLOEvaluator,
	metricsExporter *metrics.MetricsExporter,
	alerter *alerts.Alerter,
	targets []config.Target,
) {
	// 1. Check all targets
	results := checker.CheckAll(ctx, targets)

	// 2. Record metrics for each check
	for _, result := range results {
		metricsExporter.RecordHealthCheck(result)
	}

	// 3. Update SLI calculations
	sliCalculator.RecordResults(results)

	// 4. Calculate current SLIs
	slis := sliCalculator.CalculateAll()

	// 5. Update SLI metrics
	for _, sli := range slis {
		metricsExporter.UpdateSLI(sli)
		log.Printf("%s - Uptime: %.2f%%, P95: %.2fms, Errors: %.2f%%",
			sli.ServiceName, sli.Uptime, sli.P95Latency, sli.ErrorRate)
	}

	// 6. Evaluate SLOs
	sloStatuses := sloEvaluator.EvaluateAll(slis)

	// 7. Update SLO metrics and send alerts
	for _, statuses := range sloStatuses {
		for _, status := range statuses {
			metricsExporter.UpdateSLOStatus(status)
			log.Printf("%s", status.Message)
		}
	}

	// 8. Process alerts
	alerter.ProcessStatuses(sloStatuses)
}

func loadTargets(filename string) ([]config.Target, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var targets []config.Target
	if err := yaml.Unmarshal(data, &targets); err != nil {
		return nil, err
	}

	return targets, nil
}

func loadSLOs(filename string) ([]config.SLO, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var slos []config.SLO
	if err := yaml.Unmarshal(data, &slos); err != nil {
		return nil, err
	}

	return slos, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"slo-monitor"}`)
}

func statusHandler(
	sliCalculator *monitor.SLICalculator,
	sloEvaluator *monitor.SLOEvaluator,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slis := sliCalculator.CalculateAll()
		statuses := sloEvaluator.EvaluateAll(slis)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(statuses); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}