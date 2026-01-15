# Monitoring Setup

This directory contains monitoring configuration for the Trading Engine platform.

## Components

- **Prometheus**: Metrics collection and storage
- **Grafana**: Metrics visualization and dashboards

## Quick Start

Start all services with docker-compose:

```bash
docker-compose up -d
```

Access monitoring interfaces:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

## Prometheus Configuration

The `prometheus.yml` file configures:
- Scrape interval: 10s globally, 5s for trading engine
- Targets: Trading engine backend at `backend:8080/metrics`

## Adding Custom Dashboards

Place dashboard JSON files in `grafana/dashboards/` directory. They will be automatically provisioned when Grafana starts.

## Troubleshooting

Check if Prometheus can reach the backend:
```bash
curl http://localhost:9090/api/v1/targets
```

View backend metrics directly:
```bash
curl http://localhost:8080/metrics
```

Note: The `/metrics` endpoint will be implemented in plan 04-04.
