package lint

import (
	"testing"
)

func TestTemplateDatasource(t *testing.T) {
	linter := NewTemplateDatasourceRule()

	for _, tc := range []struct {
		result    Result
		dashboard Dashboard
	}{
		// 0 Data Sources
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' does not have a templated data source",
			},
			dashboard: Dashboard{
				Title: "test",
			},
		},
		// 1 Data Source
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' templated data source variable named 'foo', should be named '_datasource', or 'datasource'",
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
				Severity: Warning,
				Message:  "Dashboard 'test' templated data source variable labeled 'bar', should be labeled 'Bar data source', or 'Data source'",
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
							Query: "bar",
							Label: "bar",
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
							Label: "Data source",
							Query: "prometheus",
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
							Label: "Prometheus data source",
							Query: "prometheus",
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
							Name:  "prometheus_datasource",
							Label: "Data source",
							Query: "prometheus",
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
							Name:  "prometheus_datasource",
							Label: "Prometheus data source",
							Query: "prometheus",
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
							Name:  "loki_datasource",
							Label: "Data source",
							Query: "loki",
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
							Label: "Data source",
							Query: "loki",
						},
					},
				},
			},
		},
		// 2 or more Data Sources
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' templated data source variable named 'datasource', should be named 'prometheus_datasource'",
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
							Label: "Data source",
							Query: "prometheus",
						},
						{
							Type:  "datasource",
							Name:  "loki_datasource",
							Label: "Data source",
							Query: "loki",
						},
						{
							Type:  "datasource",
							Name:  "influx_datasource",
							Label: "Data source",
							Query: "influx",
						},
					},
				},
			},
		},
		{
			result: Result{
				Severity: Warning,
				Message:  "Dashboard 'test' templated data source variable labeled 'Data source', should be labeled 'Prometheus data source'",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Name:  "prometheus_datasource",
							Label: "Data source",
							Query: "prometheus",
						},
						{
							Type:  "datasource",
							Name:  "loki_datasource",
							Label: "Data source",
							Query: "loki",
						},
						{
							Type:  "datasource",
							Name:  "influx_datasource",
							Label: "Data source",
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
							Name:  "prometheus_datasource",
							Label: "Prometheus data source",
							Query: "prometheus",
						},
						{
							Type:  "datasource",
							Name:  "loki_datasource",
							Label: "Loki data source",
							Query: "loki",
						},
						{
							Type:  "datasource",
							Name:  "influx_datasource",
							Label: "Influx data source",
							Query: "influx",
						},
					},
				},
			},
		},
	} {
		testRule(t, linter, tc.dashboard, tc.result)
	}
}
