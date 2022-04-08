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
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' has 0 templated datasources, should be 1 or 2",
			},
			dashboard: Dashboard{
				Title: "test",
			},
		},
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' templated datasource variable named 'foo', should be named 'datasource'",
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
		// multiple datasources cases below
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
					},
				},
			},
		},
		// swap
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
							Label: "Loki Data Source",
							Query: "loki",
						},
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
				Severity: Error,
				Message:  "Dashboard 'test' with 2 templated datasources should have 'prometheus' and 'loki' types",
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
				Severity: Error,
				Message:  "Dashboard 'test' templated datasource variable named 'logs_datasource', should be named 'loki_datasource'",
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
							Name:  "logs_datasource",
							Label: "Loki Data Source",
							Query: "loki",
						},
					},
				},
			},
		},
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' templated datasource variable labeled 'Logs Data Source', should be labeled 'Loki Data Source'",
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
							Label: "Logs Data Source",
							Query: "loki",
						},
					},
				},
			},
		},
	} {
		testRule(t, linter, tc.dashboard, tc.result)
	}
}
