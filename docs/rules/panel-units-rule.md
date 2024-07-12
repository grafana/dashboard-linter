# panel-units-rule
Checks that every panel has a unit specified, and that the unit is valid per the [current list](https://github.com/grafana/grafana/blob/main/packages/grafana-data/src/valueFormats/categories.ts) defined in Grafana.

It currently only checks panels of type ["stat", "singlestat", "graph", "table", "timeseries", "gauge"].

# Best Practice
All panels should have all of their axis labeled with an apprioriate unit.

# Possible exceptions
A panel may be visualizing something which does not have a predefined unit, or which is self explanatory from the vizualization title. In this case you may wish to create a lint exclusion for this rule.