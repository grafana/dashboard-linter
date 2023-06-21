package lint

import (
	"testing"
)

func TestPanelNoTargets(t *testing.T) {
	linter := NewPanelNoTargetsRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' has no targets",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
			},
		},
		{
			result: ResultSuccess,
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
				Targets: []Target{
					{
						Expr: `sum(rate(foo[5m]))`,
					},
				},
			},
		},
	} {
		testRule(t, linter, Dashboard{Title: "test", Panels: []Panel{tc.panel}}, tc.result)
	}
}
