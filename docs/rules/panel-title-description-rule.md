# panel-title-description-rule
Checks that every panel has a title and description.

It currently only checks panels of type ["stat", "singlestat", "graph", "table", "timeseries", "gauge"].

# Best Practice
All panels should always have a title which clearly describes the panels purpose.

All panels should also have a more detailed description which appears in the tooltip for the panel.

# Possible exceptions
If a panel is sufficiently descriptive in it's title and visualization, you may wish to exclude a description and create a lint exclusion for this rule.