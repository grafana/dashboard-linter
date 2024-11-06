# template-required-variables-rule
Checks that each dashboard has a templated variable based on provided rule settings and detected variable usage for the target-required-matchers-rule.

# Best Practice
The rule ensures all of the following conditions.

* The dashboard template exists.
* The dashboard template is named `xxx`.
* The dashboard template is labeled `xxx`.
* The dashboard template uses a templated datasource, specifically named `$datasource`.
* The dashboard template uses a Prometheus query to find available matching instances.
* The dashboard template is multi select
* The dashboard template has an allValue of `.+`

## Rule Settings

```yaml
settings:
  template-required-variables-rule:
    variables:
      - cluster
      - namespace
```
Legacy config example for job and instance
```yaml
settings:
  template-required-variables-rule:
    variables:
      - job
      - instance
```