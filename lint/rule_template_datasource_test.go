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
				Severity: Error,
				Message:  "Dashboard 'test' templated data source variable labeled 'bar', should be labeled 'Bar Data Source', or 'Data Source'",
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
							Label: "Data Source",
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
							Label: "Prometheus Data Source",
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
							Label: "Data Source",
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
							Label: "Prometheus Data Source",
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
							Label: "Data Source",
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
							Label: "Data Source",
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
							Label: "Data Source",
							Query: "prometheus",
						},
						{
							Type:  "datasource",
							Name:  "loki_datasource",
							Label: "Data Source",
							Query: "loki",
						},
						{
							Type:  "datasource",
							Name:  "influx_datasource",
							Label: "Data Source",
							Query: "influx",
						},
					},
				},
			},
		},
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' templated data source variable labeled 'Data Source', should be labeled 'Prometheus Data Source'",
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
							Label: "Data Source",
							Query: "prometheus",
						},
						{
							Type:  "datasource",
							Name:  "loki_datasource",
							Label: "Data Source",
							Query: "loki",
						},
						{
							Type:  "datasource",
							Name:  "influx_datasource",
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
							Name:  "prometheus_datasource",
							Label: "Prometheus Data Source",
							Query: "prometheus",
						},
						{
							Type:  "datasource",
							Name:  "loki_datasource",
							Label: "Loki Data Source",
							Query: "loki",
						},
						{
							Type:  "datasource",
							Name:  "influx_datasource",
							Label: "Influx Data Source",
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
