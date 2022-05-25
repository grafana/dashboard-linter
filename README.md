# Grafana Dashboard Linter

This is a tool to lint Grafana dashboards for common mistakes.  To use:

```
$ go install github.com/grafana/dashboard-linter
$ dashboard-linter lint dashboard.json
```

This tool is a work in progress, and its very early days.  Right now its focused exclusively on dashboards that use a Prometheus datasource.

## Rules

The linter implements the following rules:

* `template-datasource-rule` - Checks that the dashboard has a templated datasource.
* `template-job-rule` - Checks that the dashboard has a templated job.
* `template-instance-rule` - Checks that the dashboard has a templated instance.
* `template-label-promql-rule` - Checks that the dashboard templated labels have proper PromQL expressions.
* `panel-datasource-rule` - Checks that each panel uses the templated datasource.
* `target-promql-rule` - Checks that each target uses a valid PromQL query.
* `target-rate-interval-rule` - Checks that each target uses $__rate_interval.
* `target-job-rule` - Checks that every PromQL query has a job matcher.
* `target-instance-rule` - Checks that every PromQL query has a instance matcher.

## Exceptions

Where the rules above don't make sense, you can drop a `.lint` file in a same directory as the dashboard telling the linter to ignore certain rules, eg:

```yaml
exclusions:
  template-job-rule:
    reason: "Job not needed, using recording rules."
```
