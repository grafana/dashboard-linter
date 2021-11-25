package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
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
							Query:      "label_values(up{job=~\"$job\"})",
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
				Message:  `Dashboard 'test', template 'namespaces' invalid PromQL query 'label_values(up{, namespace)': 1:17: parse error: unexpected "," in label matching, expected identifier or "}"`,
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
	} {
		require.Equal(t, tc.result, linter.LintDashboard(tc.dashboard))
	}
}
