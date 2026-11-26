# panel-datasource-rule
This rule checks each panel to be sure that it is using a templated datasource.

It currently only checks panels of type ["singlestat", "graph", "table", "timeseries"].

Panels that use one of Grafana's built-in datasources, `-- Dashboard --`, `-- Mixed --` or `-- Grafana --`, are not reported, since a datasource variable cannot stand in for them.
