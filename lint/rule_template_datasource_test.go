package lint

import (
	"testing"
)

func TestTemplateDatasource(t *testing.T) {
	linter := NewTemplateDatasourceRule()

	for _, tc := range []struct {
		name      string
		result    []Result
		dashboard Dashboard
	}{
		// 0 Data Sources
		{
			name: "0 Data Sources",
			result: []Result{{
				Severity: Error,
				Message:  "Dashboard 'test' does not have a templated data source",
			}},
			dashboard: Dashboard{
				Title: "test",
			},
		},
		// 1 Data Source
		{
			name: "1 Data Source",
			result: []Result{
				{
					Severity: Error,
					Message:  "Dashboard 'test' templated data source variable named 'foo', should be named '_datasource', or 'datasource'",
				},
				{
					Severity: Warning,
					Message:  "Dashboard 'test' templated data source variable labeled '', should be labeled ' data source', or 'Data source'",
				},
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
			name: "wrong name",
			result: []Result{
				{
					Severity: Warning,
					Message:  "Dashboard 'test' templated data source variable labeled 'bar', should be labeled 'Bar data source', or 'Data source'",
				},
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
			name:   "OK - Data source ",
			result: []Result{ResultSuccess},
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
			name:   "OK - Prometheus data source",
			result: []Result{ResultSuccess},
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
			name:   "OK - name: prometheus_datasource",
			result: []Result{ResultSuccess},
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
			name:   "OK - name: prometheus_datasource, label: Prometheus data source",
			result: []Result{ResultSuccess},
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
			name:   "OK - name: loki_datasource, query: loki",
			result: []Result{ResultSuccess},
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
			name:   "OK - name: datasource, query: loki",
			result: []Result{ResultSuccess},
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
			name: "3 Data Sources - 0",
			result: []Result{
				{
					Severity: Error,
					Message:  "Dashboard 'test' templated data source variable named 'datasource', should be named 'prometheus_datasource'",
				},
				{
					Severity: Warning,
					Message:  "Dashboard 'test' templated data source variable labeled 'Data source', should be labeled 'Prometheus data source'",
				},
				{
					Severity: Warning,
					Message:  "Dashboard 'test' templated data source variable labeled 'Data source', should be labeled 'Loki data source'",
				},
				{
					Severity: Warning,
					Message:  "Dashboard 'test' templated data source variable labeled 'Data source', should be labeled 'Influx data source'",
				},
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
			name: "3 Data Sources - 1",
			result: []Result{
				{
					Severity: Warning,
					Message:  "Dashboard 'test' templated data source variable labeled 'Data source', should be labeled 'Prometheus data source'",
				},
				{
					Severity: Warning,
					Message:  "Dashboard 'test' templated data source variable labeled 'Data source', should be labeled 'Loki data source'",
				},
				{
					Severity: Warning,
					Message:  "Dashboard 'test' templated data source variable labeled 'Data source', should be labeled 'Influx data source'",
				},
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
			name:   "3 Data Sources - 2",
			result: []Result{ResultSuccess},
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
		t.Run(tc.name, func(t *testing.T) {
			testMultiResultRule(t, linter, tc.dashboard, tc.result)
		})
	}
}
