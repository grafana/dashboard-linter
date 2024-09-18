# target-logql-auto-rule

This rule ensures that all Loki queries in a dashboard use the `$__auto` variable for range vector selectors. Using `$__auto` allows for dynamic adjustment of the range based on the dashboard's time range and resolution, providing more accurate and performant queries.

## Best Practice

Using `$__auto` instead of hard-coded time ranges like `[5m]` provides several benefits:

1. **Consistency**: It ensures a consistent approach across all Loki queries in the dashboard.
2. **Query type optimization**: It correctly uses `$__interval` for "Range" queries and `$__range` for "Instant" queries, optimizing the query for the specific type being used.
3. **Versatility**: The `$__auto` variable is automatically substituted with the step value for range queries, and with the selected time range's value (computed from the starting and ending times) for instant queries, making it suitable for various query types.

A detailed explanation can be found in the [Grafana Cloud documentation](https://grafana.com/docs/grafana-cloud/connect-externally-hosted/data-sources/loki/template-variables/#use-__auto-variable-for-loki-metric-queries).

### Examples

#### Invalid

```logql
sum(count_over_time({job="mysql"} |= "duration" [5m]))
```

#### Valid

```logql
sum(count_over_time({job="mysql"} |= "duration" [$__auto]))
```

## Possible exceptions

There may be cases where a specific, fixed time range is required for a particular query. In such cases, you may wish to create a [lint exclusion](../index.md#exclusions-and-warnings) for this rule.
