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
  -h, --help      help for lint
      --strict    fail upon linting error or warning
      --verbose   show more information about linting
```

# Rules

The linter implements the following rules:

* [template-datasource-rule](./docs/rules/template-datasource-rule.md) - Checks that the dashboard has a templated datasource.
* `template-job-rule` - Checks that the dashboard has a templated job.
* `template-instance-rule` - Checks that the dashboard has a templated instance.
* `template-label-promql-rule` - Checks that the dashboard templated labels have proper PromQL expressions.
* `panel-datasource-rule` - Checks that each panel uses the templated datasource.
* `panel-title-description-rule` - Checks that each panel has a title and description.
* `panel-units-rule` - Checks that each panel uses has valid units defined.
* `target-promql-rule` - Checks that each target uses a valid PromQL query.
* `target-rate-interval-rule` - Checks that each target uses $__rate_interval.
* `target-job-rule` - Checks that every PromQL query has a job matcher.
* `target-instance-rule` - Checks that every PromQL query has a instance matcher.

## Related Rules

There are groups of rules that are intended to drive certain outcomes, but may be implemented separately to allow more granular [exceptions](#exclusions-and-warnings), and to keep the rules terse.

## Job and Instance Template Variables

The following rules work together to ensure that every dashboard has template variables for Job and Instance, and that they are properly configured, and used in every promql query.

* `template-job-rule`
* `template-instance-rule`
* `target-job-rule`
* `target-instance-rule`

The reasoning is that.. WIP, but...
* All prom metrics will have these, as it's part of the spec
* Other reasons?

# Exclusions and Warnings

Where the rules above don't make sense, you can drop a `.lint` file in a same directory as the dashboard telling the linter to ignore certain rules, or downgrade them to warnings.

Example:
```yaml
exclusions:
  template-job-rule:
warnings:
  template-instance-rule:
```

## Reasons

It is advised that whenever you exclude, or warn for a rule, that you provide a reason. This allows for other maintainers of your dashboard to understand why a particular rule may not be followed. Eventually the dashboard-linter will provide reporting that echos that reason back to the user.

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
