package lint

import (
	"testing"
)

func TestTemplateLabelPromQLRule(t *testing.T) {
	linter := NewTemplateLabelPromQLRule()

	for _, tc := range []struct {
		result    Result
		dashboard Dashboard
	}{
		// Don't fail on non prometheus template.
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
							Query: "foo",
						},
					},
				},
			},
		},
		// What success looks like.
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
							Query: "prometheus",
						},
						{
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "label_values(up{job=~\"$job\"}, namespace)",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
		// What failure looks like.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'test', template 'namespaces' invalid templated label 'label_values(up{, namespace)': 1:4: parse error: unexpected "," in label matching, expected identifier or "}"`,
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "label_values(up{, namespace)",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
		// Invalid function.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'test', template 'namespaces' invalid templated label 'foo(up, namespace)': invalid 'function': foo`,
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "foo(up, namespace)",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
		// Invalid query expression.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'test', template 'namespaces' invalid templated label 'foo': invalid 'query': foo`,
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "foo",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
		// Support main grafana variables.
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
							Query: "prometheus",
						},
						{
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "query_result(max by(namespaces) (max_over_time(memory{}[$__range])))",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
	} {
		testRule(t, linter, tc.dashboard, tc.result)
	}
}
