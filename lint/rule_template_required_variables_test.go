package lint

import (
	"testing"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/pkg/labels"
)

func TestTemplateRequiredVariable(t *testing.T) {
	linter := NewTemplateRequiredVariablesRule(
		&TemplateRequiredVariablesRuleSettings{
			Variables: []string{"job"},
		},
		&TargetRequiredMatchersRuleSettings{
			Matchers: config.Matchers{
				{
					Name:  "instance",
					Type:  labels.MatchRegexp,
					Value: "$instance",
				},
			},
		})

	for _, tc := range []struct {
		name      string
		result    []Result
		dashboard Dashboard
	}{
		{
			name:   "Non-promtheus dashboards shouldn't fail.",
			result: []Result{ResultSuccess},
			dashboard: Dashboard{
				Title: "test",
			},
		},
		{
			name: "Missing job/instance template.",
			result: []Result{
				{Severity: Error, Message: "Dashboard 'test' is missing the job template"},
				{Severity: Error, Message: "Dashboard 'test' is missing the instance template"},
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
		{
			name: "Wrong datasource.",
			result: []Result{
				{Severity: Error, Message: "Dashboard 'test' job template should use datasource '$datasource', is currently 'foo'"},
				{Severity: Error, Message: "Dashboard 'test' job template should be a Prometheus query, is currently ''"},
				{Severity: Warning, Message: "Dashboard 'test' job template should be a labeled 'Job', is currently ''"},
				{Severity: Error, Message: "Dashboard 'test' job template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' job template allValue should be '.+', is currently ''"},
				{Severity: Error, Message: "Dashboard 'test' instance template should use datasource '$datasource', is currently 'foo'"},
				{Severity: Error, Message: "Dashboard 'test' instance template should be a Prometheus query, is currently ''"},
				{Severity: Warning, Message: "Dashboard 'test' instance template should be a labeled 'Instance', is currently ''"},
				{Severity: Error, Message: "Dashboard 'test' instance template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' instance template allValue should be '.+', is currently ''"},
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
							Name:       "job",
							Datasource: "foo",
						},
						{
							Name:       "instance",
							Datasource: "foo",
						},
					},
				},
			},
		},
		{
			name: "Wrong type.",
			result: []Result{
				{Severity: Error, Message: "Dashboard 'test' job template should be a Prometheus query, is currently 'bar'"},
				{Severity: Warning, Message: "Dashboard 'test' job template should be a labeled 'Job', is currently ''"},
				{Severity: Error, Message: "Dashboard 'test' job template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' job template allValue should be '.+', is currently ''"},
				{Severity: Error, Message: "Dashboard 'test' instance template should be a Prometheus query, is currently 'bar'"},
				{Severity: Warning, Message: "Dashboard 'test' instance template should be a labeled 'Instance', is currently ''"},
				{Severity: Error, Message: "Dashboard 'test' instance template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' instance template allValue should be '.+', is currently ''"},
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
							Name:       "job",
							Datasource: "$datasource",
							Type:       "bar",
						},
						{
							Name:       "instance",
							Datasource: "$datasource",
							Type:       "bar",
						},
					},
				},
			},
		},
		{
			name: "Wrong job/instance label.",
			result: []Result{
				{Severity: Warning, Message: "Dashboard 'test' job template should be a labeled 'Job', is currently 'bar'"},
				{Severity: Error, Message: "Dashboard 'test' job template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' job template allValue should be '.+', is currently ''"},
				{Severity: Warning, Message: "Dashboard 'test' instance template should be a labeled 'Instance', is currently 'bar'"},
				{Severity: Error, Message: "Dashboard 'test' instance template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' instance template allValue should be '.+', is currently ''"},
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
							Name:       "job",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "bar",
						},
						{
							Name:       "instance",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "bar",
						},
					},
				},
			},
		},
		{
			name:   "OK",
			result: []Result{ResultSuccess},
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
							Name:       "job",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "Job",
							Multi:      true,
							AllValue:   ".+",
						},
						{
							Name:       "instance",
							Datasource: "${datasource}",
							Type:       "query",
							Label:      "Instance",
							Multi:      true,
							AllValue:   ".+",
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

func TestTemplateRequiredVariableNilRequiredMatchers(t *testing.T) {
	linter := NewTemplateRequiredVariablesRule(
		nil,
		&TargetRequiredMatchersRuleSettings{
			Matchers: config.Matchers{
				{
					Name:  "instance",
					Type:  labels.MatchRegexp,
					Value: "$instance",
				},
			},
		})

	for _, tc := range []struct {
		name      string
		result    []Result
		dashboard Dashboard
	}{
		{
			name: "Missing instance template.",
			result: []Result{
				{Severity: Error, Message: "Dashboard 'test' is missing the instance template"},
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
							Name:       "job",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "Job",
							Multi:      true,
							AllValue:   ".+",
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

func TestTemplateRequiredVariableNilConfig(t *testing.T) {
	linter := NewTemplateRequiredVariablesRule(
		&TemplateRequiredVariablesRuleSettings{
			Variables: []string{"job"},
		},
		nil)

	for _, tc := range []struct {
		name      string
		result    []Result
		dashboard Dashboard
	}{
		{
			name: "Missing job template.",
			result: []Result{
				{Severity: Error, Message: "Dashboard 'test' is missing the job template"},
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
							Name:       "instance",
							Datasource: "${datasource}",
							Type:       "query",
							Label:      "Instance",
							Multi:      true,
							AllValue:   ".+",
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

func TestTemplateRequiredVariableNilInput(t *testing.T) {
	linter := NewTemplateRequiredVariablesRule(
		nil,
		nil)

	for _, tc := range []struct {
		name      string
		result    []Result
		dashboard Dashboard
	}{
		{
			name:   "OK",
			result: []Result{ResultSuccess},
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
							Name:       "job",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "Job",
							Multi:      true,
							AllValue:   ".+",
						},
						{
							Name:       "instance",
							Datasource: "${datasource}",
							Type:       "query",
							Label:      "Instance",
							Multi:      true,
							AllValue:   ".+",
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
