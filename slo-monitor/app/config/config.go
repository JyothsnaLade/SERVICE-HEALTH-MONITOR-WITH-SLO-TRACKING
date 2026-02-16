package config

import (
	"os"
	"time"
)

type Config struct {
	CheckInterval    time.Duration
	PrometheusPort   string
	TargetsFile      string
	SLOsFile         string
	AlertWebhookURL  string
}

type Target struct {
	Name     string   `yaml:"name"`
	URL      string   `yaml:"url"`
	Method   string   `yaml:"method"`
	Timeout  int      `yaml:"timeout"`
	Expected int      `yaml:"expected_status"`
	Tags     []string `yaml:"tags"`
}

type SLO struct {
	Name              string  `yaml:"name"`
	Service           string  `yaml:"service"`
	UptimeTarget      float64 `yaml:"uptime_target"`      
	LatencyTarget     int     `yaml:"latency_target"`     
	ErrorBudget       float64 `yaml:"error_budget"`       
	EvaluationWindow  string  `yaml:"evaluation_window"` 
}

func Load() *Config {
	checkInterval, _ := time.ParseDuration(getEnv("CHECK_INTERVAL", "30s"))
	
	return &Config{
		CheckInterval:   checkInterval,
		PrometheusPort:  getEnv("PROMETHEUS_PORT", "9090"),
		TargetsFile:     getEnv("TARGETS_FILE", "/config/targets.yaml"),
		SLOsFile:        getEnv("SLOS_FILE", "/config/slos.yaml"),
		AlertWebhookURL: getEnv("ALERT_WEBHOOK_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}