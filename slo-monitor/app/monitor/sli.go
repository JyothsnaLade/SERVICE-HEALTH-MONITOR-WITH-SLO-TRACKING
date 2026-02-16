package monitor

import (
	"sync"
	"time"
)

type SLI struct {
	ServiceName    string
	Uptime         float64
	AvgLatency     float64
	P95Latency     float64
	P99Latency     float64
	ErrorRate      float64
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	LastUpdated    time.Time
}

type SLICalculator struct {
	mu      sync.RWMutex
	results map[string][]CheckResult
	window  time.Duration
}

func NewSLICalculator(window time.Duration) *SLICalculator {
	return &SLICalculator{
		results: make(map[string][]CheckResult),
		window:  window,
	}
}

func (sc *SLICalculator) RecordResults(results []CheckResult) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sc.window)

	for _, result := range results {
		serviceName := result.Target.Name
		
		sc.results[serviceName] = append(sc.results[serviceName], result)
		
		filtered := make([]CheckResult, 0)
		for _, r := range sc.results[serviceName] {
			if r.Timestamp.After(cutoff) {
				filtered = append(filtered, r)
			}
		}
		sc.results[serviceName] = filtered
	}
}

func (sc *SLICalculator) Calculate(serviceName string) SLI {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	results, exists := sc.results[serviceName]
	if !exists || len(results) == 0 {
		return SLI{ServiceName: serviceName}
	}

	sli := SLI{
		ServiceName: serviceName,
		LastUpdated: time.Now(),
	}

	var totalLatency time.Duration
	latencies := make([]float64, 0, len(results))

	for _, r := range results {
		sli.TotalRequests++
		
		if r.Success {
			sli.SuccessCount++
		} else {
			sli.FailureCount++
		}
		
		totalLatency += r.Latency
		latencies = append(latencies, float64(r.Latency.Milliseconds()))
	}

	sli.Uptime = (float64(sli.SuccessCount) / float64(sli.TotalRequests)) * 100
	sli.ErrorRate = (float64(sli.FailureCount) / float64(sli.TotalRequests)) * 100
	sli.AvgLatency = float64(totalLatency.Milliseconds()) / float64(len(results))

	if len(latencies) > 0 {
		sli.P95Latency = percentile(latencies, 95)
		sli.P99Latency = percentile(latencies, 99)
	}

	return sli
}

func percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	sorted := make([]float64, len(data))
	copy(sorted, data)
	
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	index := int((p / 100.0) * float64(len(sorted)))
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	
	return sorted[index]
}

func (sc *SLICalculator) CalculateAll() map[string]SLI {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	slis := make(map[string]SLI)
	
	for serviceName := range sc.results {
		slis[serviceName] = sc.Calculate(serviceName)
	}
	
	return slis
}
