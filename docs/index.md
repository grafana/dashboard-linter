# Usage
All Commands:

[embedmd]:# (_intermediate/help.txt)
```txt
A command-line application to lint Grafana dashboards.

Usage:
  dashboard-linter [flags]
  dashboard-linter [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  lint        Lint a dashboard
  rules       Print documentation about each lint rule.

Flags:
  -h, --help   help for dashboard-linter

Use "dashboard-linter [command] --help" for more information about a command.
```

## Completion
[embedmd]:# (_intermediate/completion.txt)
```txt
Generate the autocompletion script for dashboard-linter for the specified shell.
See each sub-command's help for details on how to use the generated script.

Usage:
  dashboard-linter completion [command]

Available Commands:
  bash        Generate the autocompletion script for bash
  fish        Generate the autocompletion script for fish
  powershell  Generate the autocompletion script for powershell
  zsh         Generate the autocompletion script for zsh

Flags:
  -h, --help   help for completion

Use "dashboard-linter completion [command] --help" for more information about a command.
```

## Lint
[embedmd]:# (_intermediate/lint.txt)
```txt
Returns warnings or errors for dashboard which do not adhere to accepted standards

Usage:
  dashboard-linter lint [dashboard.json] [flags]

Flags:
  -c, --config string   path to a configuration file
      --experimental    enable experimental rules
      --fix             automatically fix problems if possible
  -h, --help            help for lint
      --stdin           read from stdin
      --strict          fail upon linting error or warning
      --verbose         show more information about linting
```

# Rules

The linter implements the following rules:

* [template-datasource-rule](./rules/template-datasource-rule) - ``stable`` - Checks that the dashboard has a templated datasource.
* [template-job-rule](./rules/template-job-rule) - `stable` - Checks that the dashboard has a templated job.
* [template-instance-rule](./rules/template-instance-rule) - `stable` - Checks that the dashboard has a templated instance.
* [template-label-promql-rule](./rules/template-label-promql-rule) - `stable` - Checks that the dashboard templated labels have proper PromQL expressions.
* [template-on-time-change-reload-rule](./rules/template-on-time-change-reload-rule) - `stable` - Checks that the dashboard template variables are configured to reload on time change.
* [panel-datasource-rule](./rules/panel-datasource-rule) - `stable` - Checks that each panel uses the templated datasource.
* [panel-title-description-rule](./rules/panel-title-description-rule) - `stable` - Checks that each panel has a title and description.
* [panel-units-rule](./rules/panel-units-rule) - `stable` - Checks that each panel uses has valid units defined.
* [panel-no-targets-rule](./rules/panel-no-targets-rule) - `stable` - Checks that each panel has at least one target.
* [target-promql-rule](./rules/target-promql-rule) - `stable` - Checks that each target uses a valid PromQL query.
* [target-rate-interval-rule](./rules/target-rate-interval-rule) - `stable` - Checks that each target uses $__rate_interval.
* [target-job-rule](./rules/target-job-rule) - `stable` - Checks that every PromQL query has a job matcher.
* [target-instance-rule](./rules/target-instance-rule) - `stable` - Checks that every PromQL query has a instance matcher.
* [target-counter-agg-rule](./rules/target-counter-agg-rule) - `stable` - Checks that any counter metric (ending in _total) is aggregated with rate, irate, or increase.
* [uneditable-dashboard-rule](./rules/uneditable-dashboard-rule) - `stable` - Checks that the dashboard is not editable.
* [target-required-matchers-rule](./rules/target-required-matchers-rule) - `experimental` - Checks that target PromQL query has the required matchers
* [template-required-variables-rule](./rules/template-required-variables-rule) - `experimental` - Checks that the dashboard has a template variable for required variables or matchers that use variables

## Rule stability
- **Stable** rules have gone through testing and been widely adopted.

- **Experimental** rules are for new and experimental features.
  These rules are not enabled by default, but can be enabled by providing the `experimental` flag.
  Allowing early adopters to gain confidence with new features.

- **Deprecated** rules may be removed or replaced when they are marked as deprecated.

## Related Rules

There are groups of rules that are intended to drive certain outcomes, but may be implemented separately to allow more granular [exceptions](#exclusions-and-warnings), and to keep the rules terse.

### Job and Instance Template Variables

The following rules work together to ensure that every dashboard has template variables for `Job` and `Instance`, that they are properly configured, and used in every promql query.

* [template-job-rule](./rules/template-job-rule.md)
* [template-instance-rule](./rules/template-instance-rule.md)
* [target-job-rule](./rules/target-job-rule.md)
* [target-instance-rule](./rules/target-instance-rule.md)

These rules enforce a best practice for dashboards with a single Prometheus or Loki data source. Metrics and logs scraped by Prometheus and Loki have automatically generated [job and instance labels](https://prometheus.io/docs/concepts/jobs_instances/) on them. For this reason, having the ability to filter by these assured always-present labels is logical and a useful additional feature.

#### Multi Data Source Exceptions
These rules may become cumbersome when dealing with a dashboard with more than one data source. Significant relabeling in the scrape config is required because the `job` and `instance` labels must match between each data source, and the default names for those labels will be different or absent in disparate data sources. 

For example:
The [Grafana Cloud Docker Integration](https://grafana.com/docs/grafana-cloud/data-configuration/integrations/integration-reference/integration-docker/#post-install-configuration-for-the-docker-integration) combines metrics from cAdvisor, and logs from the docker daemon using `docker_sd_configs`.

In this case, without label rewriting, the logs would not have any labels at all. The metrics relabeling applies opinionated job names rather than the defaults provided by the agent. (`integrations/cadvisor`).

For dashboards like this, create a linting [exception](#exclusions-and-warnings) for these rules, and use a separate label that exists on data from all data sources to filter.

# Exclusions and Warnings

Where the rules above don't make sense, you can add a `.lint` file in the same directory as the dashboard telling the linter to ignore certain rules or downgrade them to a warning.

Example:
```yaml
exclusions:
  template-job-rule:
warnings:
  template-instance-rule:
```

## Reasons

Whenever you exclude or warn for a rule, it's recommended that you provide a reason. This allows for other maintainers of your dashboard to understand why a particular rule may not be followed. Eventually, the dashboard-linter will provide reporting that echoes that reason back to the user.

Example:
```yaml
exclusions:
  template-job-rule:
    reason: A job matcher is hardcoded into the recording rule used for all queries on these dashboards.
```

## Multiple Entries and Specific Exclusions

It is possible to not exclude for every violation of a rule. Whenever possible, it is advised that you exclude *only* the rule violations that are necessary, and that you specifically identify them along with a reason. This will allow the linter to catch the same rule violation, which may happen on another dashboard, panel, or target when modifications are made.

Example:
```yaml
exclusions:
  target-rate-interval-rule:
    reason: Top 10's are intended to be displayed for the currently selected range.
    entries:
    - dashboard: Apollo Server
      panel: Top 10 Duration Rate
    - dashboard: Apollo Server
      panel: Top 10 Slowest Fields Resolution
  target-instance-rule:
    reason: Totals are intended to be across all instances
    entries:
    - panel: Requests Per Second
      targetIdx: 2
    - panel: Response Latency
      targetIdx: 2
```
