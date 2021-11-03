package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateDatasource(t *testing.T) {
	linter := NewTemplateDatasourceRule()

	for _, tc := range []struct {
		result    Result
		dashboard Dashboard
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' does not have a templated datasource",
			},
			dashboard: Dashboard{
				Title: "test",
			},
		},
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' templated datasource variable named 'foo', should be names 'datasource'",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type: "datasource",
							Name: "foo",
						},
					},
				},
			},
		},
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' templated datasource variable labeled 'bar', should be labeled 'Data Source'",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Name:  "datasource",
							Label: "bar",
						},
					},
				},
			},
		},
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' templated datasource variable query is 'influx', should be 'prometheus' or 'loki'",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Name:  "datasource",
							Label: "Data Source",
							Query: "influx",
						},
					},
				},
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Name:  "datasource",
							Label: "Data Source",
							Query: "prometheus",
						},
					},
				},
			},
		},
	} {
		require.Equal(t, tc.result, linter.LintDashboard(tc.dashboard))
	}
}
