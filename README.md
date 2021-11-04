# Grafana Dashboard Linter

This is a tool to lint Grafana dashboards for common mistakes.  To use:

```
$ go install github.com/grafana/dashboard-linter
$ dashboard-linter lint dashboard.json
```

This tool is a work in progress, and its very early days.  Right now its focused exclusively on dashboards that use a Prometheus datasource.