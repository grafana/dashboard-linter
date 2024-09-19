# target-required-matchers-rule
Checks that each PromQL query has a the matchers specified in rule settings. This rule is experimental and is designed to work with Prometheus datasources.

## Rule Settings

```yaml
settings:
  target-required-matchers-rule:
    matchers:
      - cluster=~"$cluster"
      - someLabel="someValue"
```
Legacy config example for job and instance
```yaml
settings:
  target-required-matchers-rule:
    matchers:
      - job
      - instance
```