package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	
	"slo-monitor/monitor"
)

type Alert struct {
	Service     string    `json:"service"`
	SLOName     string    `json:"slo_name"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Uptime      float64   `json:"uptime"`
	ErrorRate   float64   `json:"error_rate"`
	P95Latency  float64   `json:"p95_latency"`
	ErrorBudget float64   `json:"error_budget_used"`
}

type Alerter struct {
	webhookURL string
	client     *http.Client
}

func NewAlerter(webhookURL string) *Alerter {
	return &Alerter{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (a *Alerter) SendAlert(status monitor.SLOStatus) error {
	if status.Compliant {
		return nil // Only alert on violations
	}

	severity := "warning"
	if status.ErrorBudgetUsed > 90 {
		severity = "critical"
	}

	alert := Alert{
		Service:     status.SLI.ServiceName,
		SLOName:     status.SLO.Name,
		Severity:    severity,
		Message:     status.Message,
		Timestamp:   status.EvaluatedAt,
		Uptime:      status.SLI.Uptime,
		ErrorRate:   status.SLI.ErrorRate,
		P95Latency:  status.SLI.P95Latency,
		ErrorBudget: status.ErrorBudgetUsed,
	}

	// Log alert
	log.Printf("ALERT [%s]: %s - %s", alert.Severity, alert.Service, alert.Message)

	// Send to webhook if configured
	if a.webhookURL != "" {
		return a.sendWebhook(alert)
	}

	return nil
}

func (a *Alerter) sendWebhook(alert Alert) error {
	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	resp, err := a.client.Post(a.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	log.Printf("Alert sent to webhook: %s", a.webhookURL)
	return nil
}

func (a *Alerter) ProcessStatuses(statuses map[string][]monitor.SLOStatus) {
	for _, serviceStatuses := range statuses {
		for _, status := range serviceStatuses {
			if err := a.SendAlert(status); err != nil {
				log.Printf("Failed to send alert: %v", err)
			}
		}
	}
}