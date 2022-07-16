package lint

import "fmt"

func NewDefaultRefreshIntervalRule() *DashboardRuleFunc {
	refreshInterval := "5m"

	return &DashboardRuleFunc{
		name:        "panel-refresh-interval-rule",
		description: "Checks that each panel has an appropriate refresh interval.",
		fn: func(d Dashboard) Result {
			if d.Refresh != refreshInterval {
				return Result{
					Severity: Error,
					Message:  fmt.Sprintf("Dashboard '%s' should have a default refresh interval of '%s', current it is: '%s'", d.Title, refreshInterval, d.Refresh),
				}
			}
			return ResultSuccess
		},
	}
}
