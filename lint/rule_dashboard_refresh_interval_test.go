package lint

import (
	"testing"
)

func TestDashboardRefreshInterval(t *testing.T) {
	linter := NewDefaultRefreshIntervalRule()
	refreshInterval := "5m"

	for _, tc := range []struct {
		result  Result
		refresh string
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' should have a default refresh interval of '" + refreshInterval + "', current it is: '60s'",
			},
			refresh: "60s",
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			refresh: "5m",
		},
	} {
		testRule(t, linter, Dashboard{Title: "test", Refresh: tc.refresh}, tc.result)
	}
}
