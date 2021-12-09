package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPanelRateIntervalRule(t *testing.T) {
	linter := NewPanelRateIntervalRule()
	dashboard := Dashboard{
		Title: "dashboard",
		Templating: struct {
			List []Template `json:"list"`
		}{
			List: []Template{
				{
					Type:  "datasource",
					Query: "prometheus",
				},
			},
		},
	}

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		// Don't fail non-prometheus panels.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Title:      "panel",
				Datasource: "foo",
			},
		},
		// This is what a valid panel looks like.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job=~"$job",instance=~"$instance"}[$__rate_interval]))`,
					},
				},
			},
		},
                // This is what a valid panel looks like.
                {
                        result: Result{
                                Severity: Success,
                                Message:  "OK",
                        },
                        panel: Panel{
                                Title: "panel",
                                Type:  "singlestat",
                                Targets: []Target{
                                        {
                                                Expr: `sum(rate(foo{job=~"$job",instance=~"$instance"}[$__rate_interval]))/sum(rate(bar{job=~"$job",instance=~"$instance"}[$__rate_interval]))`,
                                        },
                                },
                        },
                },
		// Invalid query
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'dashboard', panel 'panel' invalid PromQL query 'sum(rate(foo{job=~"$job",instance=~"$instance"}[5m]))': should use $__rate_interval`,
			},
			panel: Panel{
				Title: "panel",
				Type:  "singlestat",
				Targets: []Target{
					{
						Expr: `sum(rate(foo{job=~"$job",instance=~"$instance"}[5m]))`,
					},
				},
			},
		},
	} {
		require.Equal(t, tc.result, linter.LintPanel(dashboard, tc.panel))
	}
}
