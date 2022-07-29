# template-datasource-rule
This rule checks that there is precisely one template variable for data source on your dashboard.

## Best Practice
The data source variable should be named `datasource` and the label should be "Data Source"

The variable may be for either a Prometheus or Loki datasource.

## Possible exceptions
Some dashboards may contain other data source types besides Prometheus or Loki.

Some dashboards may contain more than one data source. This rule will be updated in the future to accomodate multiple data sources.