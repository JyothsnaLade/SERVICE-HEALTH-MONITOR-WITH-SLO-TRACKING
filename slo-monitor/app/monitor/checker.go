package monitor

import (
	"context"
	"net/http"
	"time"
	
	"slo-monitor/config"
)

type CheckResult struct {
	Target       config.Target
	Success      bool
	StatusCode   int
	Latency      time.Duration
	Error        error
	Timestamp    time.Time
}

type HealthChecker struct {
	client *http.Client
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (hc *HealthChecker) Check(ctx context.Context, target config.Target) CheckResult {
	result := CheckResult{
		Target:    target,
		Timestamp: time.Now(),
	}

	timeout := time.Duration(target.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	method := target.Method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(ctx, method, target.URL, nil)
	if err != nil {
		result.Error = err
		return result
	}

	start := time.Now()
	resp, err := hc.client.Do(req)
	result.Latency = time.Since(start)

	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	
	expectedStatus := target.Expected
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	result.Success = resp.StatusCode == expectedStatus

	return result
}

func (hc *HealthChecker) CheckAll(ctx context.Context, targets []config.Target) []CheckResult {
	results := make([]CheckResult, len(targets))
	
	for i, target := range targets {
		results[i] = hc.Check(ctx, target)
	}
	
	return results
}
