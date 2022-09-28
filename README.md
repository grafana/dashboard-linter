# Grafana Dashboard Linter

This tool is a command-line application to lint Grafana dashboards for common mistakes, and suggest best practices. To use the linter, run the following install commands:

```
$ go install github.com/grafana/dashboard-linter@latest
$ dashboard-linter lint dashboard.json
```

This tool is a work in progress and it's still very early days. The current capabilities are focused exclusively on dashboards that use a Prometheus data source.

See [the docs](docs/index.md) for more detail.
