# template-variable-matchers-rule
Checks that each dashboard has a templated variable corresponding to a required matcher variable. See [Required Matchers And Template Variables](../index.md#required-matchers-and-template-variables) for more information about rules relating to this one.

# Best Practice
The rule ensures all of the following conditions.

* The dashboard template exists.
* The dashboard template is named after the required variable by matcher.
* The dashboard template is labeled after the required variable by matcher.
* The dashboard template uses a templated datasource, specifically named `$datasource`.
* The dashboard template uses a Prometheus query to find available matching instances.
* The dashboard template is multi select
* The dashboard template has an allValue of `.+`

