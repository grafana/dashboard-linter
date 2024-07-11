# template-job-rule
Checks that each dashboard has a templated job. See [Job and Instance Template Variables](../index.md#job-and-instance-template-variables) for more information about rules relating to this one.

# Best Practice
The rule ensures all of the following conditions.

* The dashboard template exists.
* The dashboard template is named `job`.
* The dashboard template is labeled `job`.
* The dashboard template uses a templated datasource, specifically named `$datasource`.
* The dashboard template uses a Prometheus query to find available matching jobs.
* The dashboard template is multi select
* The dashboard template has an allValue of `.+`

