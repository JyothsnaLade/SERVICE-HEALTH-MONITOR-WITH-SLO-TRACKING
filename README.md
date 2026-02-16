# Service Health Monitor with SLO Tracking

Production-grade SLO monitoring system that tracks service health across multiple endpoints, calculates Service Level Indicators (SLIs), evaluates against Service Level Objectives (SLOs), and provides real-time alerting on violations.

## Features

- **Multi-service monitoring** - Monitor any HTTP endpoint
- **SLI calculation** - Uptime, latency (avg, P95, P99), error rate
- **SLO evaluation** - Compare SLIs against targets
- **Error budget tracking** - Monitor budget consumption
- **Prometheus metrics** - Industry-standard metrics export
- **Grafana dashboards** - Beautiful visualizations
- **Alert system** - Webhook integration for violations

Production Ready

- Health checks and readiness probes
- Graceful shutdown handling
- Configurable via YAML files
- Docker Compose orchestration
- 24-hour rolling window analysis
- Automatic service discovery

## Prerequisites

- Docker
- Docker Compose
- (Optional) Go 1.21+ for local development

## Grafana Dashboard
![Grafana Dashboard](/slo-monitor/docs/screenshots/grafana-dashboard.png)

## Prometheus Targets
![Prometheus Targets](/slo-monitor/docs/screenshots/prometheus-targets.png)

## Docker Services Running
![Docker Services Running](/slo-monitor/docs/screenshots/Docker-status.png)

## SLO Monitor Logs
![SLO Monitor Logs](/slo-monitor/docs/screenshots/slo-monitor-logs.png)

## Project Structure
![Project Structure](/slo-monitor/docs/screenshots/project-structure.png)


## Quick Start

### 1. Clone and Setup
```bash
# Clone repository
git clone https://github.com/YOUR-USERNAME/slo-monitor.git
cd slo-monitor

# Verify structure
ls -la
```

### 2. Start the Stack
```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f slo-monitor
```

### 3. Access Services

- **SLO Monitor**: http://localhost:9090
  - Metrics: http://localhost:9090/metrics
  - Health: http://localhost:9090/health
  - Status: http://localhost:9090/status
  
- **Prometheus**: http://localhost:9091
  - Query metrics and create alerts
  
- **Grafana**: http://localhost:3000
  - Username: `admin`
  - Password: `admin`
  - Dashboard: "SLO Monitor Dashboard"

### 4. View Results

**Check SLO Monitor logs:**
```bash
docker-compose logs -f slo-monitor
```

**View Grafana Dashboard:**
1. Open http://localhost:3000
2. Login (admin/admin)
3. Go to "Dashboards" → "SLO Monitor Dashboard"
4. See live metrics!

## Prometheus Metrics Exposed

### Health Check Metrics
- `slo_health_check_total{service,status}` - Total checks performed
- `slo_health_check_duration_seconds{service}` - Check duration

### SLI Metrics
- `slo_service_uptime_percent{service}` - Current uptime %
- `slo_service_latency_avg_milliseconds{service}` - Average latency
- `slo_service_latency_p95_milliseconds{service}` - P95 latency
- `slo_service_latency_p99_milliseconds{service}` - P99 latency
- `slo_service_error_rate_percent{service}` - Error rate %
- `slo_service_total_requests{service}` - Total requests

### SLO Compliance Metrics
- `slo_compliance{service,slo_name}` - 1=compliant, 0=violated
- `slo_uptime_target_percent{service,slo_name}` - Uptime target
- `slo_latency_target_milliseconds{service,slo_name}` - Latency target
- `slo_error_budget_used_percent{service,slo_name}` - Budget consumed

## Alerts

### Configure Webhook

Set environment variable in `docker-compose.yml`:
```yaml
environment:
  - ALERT_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Alert Format
```json
{
  "service": "google",
  "slo_name": "google-slo",
  "severity": "warning",
  "message": "SLO violation - [Uptime: 98.50% < 99.90%]",
  "timestamp": "2024-02-14T20:30:00Z",
  "uptime": 98.5,
  "error_rate": 1.5,
  "p95_latency": 523.4,
  "error_budget_used": 150.0
}
```


## Cleanup
```bash
# Stop all services
docker-compose down

# Remove volumes (deletes metrics data)
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

## Troubleshooting

### Service Won't Start
```bash
# Check logs
docker-compose logs slo-monitor

# Common issues:
# - Config files not mounted: Check volumes in docker-compose.yml
# - Port already in use: Change ports in docker-compose.yml
```

### No Metrics in Grafana
```bash
# Check Prometheus is scraping
# Go to http://localhost:9091/targets
# Verify slo-monitor endpoint is UP

# Check Grafana datasource
# Go to http://localhost:3000/datasources
# Test Prometheus connection
```

### SLO Always Violated
```bash
# Check if service is actually up
curl -v https://your-service.com

# Adjust SLO targets if too strict
# Edit config/slos.yaml and increase targets
```

## Resources

- [SLO Best Practices](https://sre.google/sre-book/service-level-objectives/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Dashboards](https://grafana.com/docs/grafana/latest/dashboards/)
- [Go Prometheus Client](https://github.com/prometheus/client_golang)


## License

MIT License

## Acknowledgments

Built as a demonstration of production-grade SLO monitoring for learning and portfolio purposes.