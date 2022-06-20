package lint

import (
	"testing"
)

func TestDashboardTimeInterval(t *testing.T) {
	linter := NewDefaultTimeIntervalRule()
	timeIntervalFrom := "now-1h"
	timeIntervalTo := "now"

	for _, tc := range []struct {
		result Result
		time   Time
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' should have a default time interval of From: '" + timeIntervalFrom + "' To: '" + timeIntervalTo + "', currently it is From: 'now-30s' To: 'now'",
			},
			time: Time{
				From: "now-30s",
				To:   "now",
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			time: Time{
				From: "now-1h",
				To:   "now",
			},
		},
	} {
		testRule(t, linter, Dashboard{Title: "test", Time: tc.time}, tc.result)
	}
}
