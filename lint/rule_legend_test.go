package lint

import (
	"testing"
)

func TestLegendRule(t *testing.T) {
	linter := NewLegendRule()

	for _, tc := range []struct {
		name   string
		result Result
		target Target
	}{
		{
			name:   "Happy path",
			result: ResultSuccess,
			target: Target{
				LegendFormat: "{{instance}} - CPU utilization",
			},
		},
		{
			name: "missing legend",
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' Legend is missing - should start with '{{instance}} - '",
			},
			target: Target{},
		},
		{
			name: "wrong legend",
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'dashboard', panel 'panel', target idx '0' Legend should start with '{{instance}} - ' - found 'CPU utilization'",
			},
			target: Target{
				LegendFormat: "CPU utilization",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dashboard := Dashboard{
				Title: "dashboard",
				Panels: []Panel{
					{
						Title:   "panel",
						Type:    "singlestat",
						Targets: []Target{tc.target},
					},
				},
			}

			testRule(t, linter, dashboard, tc.result)
		})
	}
}
