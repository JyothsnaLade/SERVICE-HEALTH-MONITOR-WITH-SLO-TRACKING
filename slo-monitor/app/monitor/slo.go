package monitor

import (
	"fmt"
	"time"
	
	"slo-monitor/config"
)

type SLOStatus struct {
	SLO              config.SLO
	SLI              SLI
	Compliant        bool
	UptimeCompliant  bool
	LatencyCompliant bool
	ErrorBudgetUsed  float64
	Message          string
	EvaluatedAt      time.Time
}

type SLOEvaluator struct {
	slos []config.SLO
}

func NewSLOEvaluator(slos []config.SLO) *SLOEvaluator {
	return &SLOEvaluator{
		slos: slos,
	}
}

func (se *SLOEvaluator) Evaluate(sli SLI) []SLOStatus {
	statuses := make([]SLOStatus, 0)
	
	for _, slo := range se.slos {
		if slo.Service != sli.ServiceName {
			continue
		}
		
		status := SLOStatus{
			SLO:         slo,
			SLI:         sli,
			EvaluatedAt: time.Now(),
		}
		
		status.UptimeCompliant = sli.Uptime >= slo.UptimeTarget
		status.LatencyCompliant = sli.P95Latency <= float64(slo.LatencyTarget)
		
		allowedErrorRate := 100.0 - slo.UptimeTarget
		actualErrorRate := sli.ErrorRate
		
		if allowedErrorRate > 0 {
			status.ErrorBudgetUsed = (actualErrorRate / allowedErrorRate) * 100
		} else {
			status.ErrorBudgetUsed = 0
		}
		
		status.Compliant = status.UptimeCompliant && status.LatencyCompliant
		
		if status.Compliant {
			status.Message = fmt.Sprintf("✅ SLO compliant - Uptime: %.2f%% (target: %.2f%%), P95 Latency: %.2fms (target: %dms)",
				sli.Uptime, slo.UptimeTarget, sli.P95Latency, slo.LatencyTarget)
		} else {
			violations := make([]string, 0)
			
			if !status.UptimeCompliant {
				violations = append(violations, 
					fmt.Sprintf("Uptime: %.2f%% < %.2f%%", sli.Uptime, slo.UptimeTarget))
			}
			
			if !status.LatencyCompliant {
				violations = append(violations, 
					fmt.Sprintf("P95 Latency: %.2fms > %dms", sli.P95Latency, slo.LatencyTarget))
			}
			
			status.Message = fmt.Sprintf("❌ SLO violation - %v", violations)
		}
		
		if status.ErrorBudgetUsed > 80 {
			status.Message += fmt.Sprintf(" ⚠️ Error budget %.1f%% consumed", status.ErrorBudgetUsed)
		}
		
		statuses = append(statuses, status)
	}
	
	return statuses
}

func (se *SLOEvaluator) EvaluateAll(slis map[string]SLI) map[string][]SLOStatus {
	allStatuses := make(map[string][]SLOStatus)
	
	for serviceName, sli := range slis {
		statuses := se.Evaluate(sli)
		if len(statuses) > 0 {
			allStatuses[serviceName] = statuses
		}
	}
	
	return allStatuses
}
