# panel-units-rule
Checks that every panel has a unit specified, and that the unit is valid per the [current list](https://github.com/grafana/grafana/blob/main/packages/grafana-data/src/valueFormats/categories.ts) defined in Grafana.

It currently only checks panels of type ["stat", "singlestat", "graph", "table", "timeseries", "gauge"].

# Best Practice
All panels should have an apprioriate unit set.

# Possible exceptions
This rule is automatically excluded when:
 - Value mappings are set in a panel.
 - A Stat panel is configured to show non-numeric values (like label's value), for that 'Fields options' are configured to any value other than 'Numeric fields' (which is default).

Also, a panel may be visualizing something which does not have a predefined unit, or which is self explanatory from the vizualization title. In this case you may wish to create a lint exclusion for this rule.