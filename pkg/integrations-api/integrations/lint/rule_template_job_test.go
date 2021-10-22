package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJobDatasource(t *testing.T) {
	linter := NewTemplateJobRule()

	for _, tc := range []struct {
		result    Result
		dashboard Dashboard
	}{
		// Non-promtheus dashboards shouldn't fail.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			dashboard: Dashboard{
				Title: "test",
			},
		},
		// Missing job template.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' is missing the job template",
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
					},
				},
			},
		},
		// Wrong job label.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' job template should be a labelled 'Job'",
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
							Name: "job",
							Type: "prometheus",
						},
					},
				},
			},
		},
		// Missing instance template.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' is missing the instance template",
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
							Name:  "job",
							Label: "Job",
							Type:  "prometheus",
						},
					},
				},
			},
		},
		// Missing instance template.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' instance template should be a labelled 'Instance'",
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
							Name:  "job",
							Label: "Job",
							Type:  "prometheus",
						},
						{
							Name: "instance",
							Type: "prometheus",
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
							Name:  "job",
							Label: "Job",
							Type:  "prometheus",
						},
						{
							Name:  "instance",
							Label: "Instance",
							Type:  "prometheus",
						},
					},
				},
			},
		},
	} {
		require.Equal(t, tc.result, linter.LintDashboard(nil, tc.dashboard))
	}
}
