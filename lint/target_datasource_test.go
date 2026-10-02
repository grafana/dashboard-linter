package lint

import (
	"fmt"
	"strings"
	"testing"
)

func dashboardFromVariables(t *testing.T, variables []Template) Dashboard {
	t.Helper()
	d, err := NewDashboard([]byte(fmt.Sprintf(`{"templating": {"list": [%s]}}`, variablesJSON(variables))))
	if err != nil {
		t.Fatalf("invalid dashboard: %v", err)
	}
	return d
}

func variablesJSON(variables []Template) string {
	parts := make([]string, 0, len(variables))
	for _, v := range variables {
		parts = append(parts, fmt.Sprintf(`{"type": %q, "name": %q, "query": %q}`, v.Type, v.Name, v.Query))
	}
	return strings.Join(parts, ",")
}

func TestGetDatasourceVariables(t *testing.T) {
	for _, tc := range []struct {
		name      string
		variables []Template
		want      map[string]string
	}{
		{
			name: "datasource variables map to their plugin id",
			variables: []Template{
				{Name: "prometheus_datasource", Type: "datasource", Query: "prometheus"},
				{Name: "loki_datasource", Type: "datasource", Query: "loki"},
			},
			want: map[string]string{
				"prometheus_datasource": "prometheus",
				"loki_datasource":       "loki",
			},
		},
		{
			name: "non datasource variables are ignored",
			variables: []Template{
				{Name: "environment", Type: "query", Query: "label_values(environment)"},
				{Name: "loki_datasource", Type: "datasource", Query: "loki"},
			},
			want: map[string]string{"loki_datasource": "loki"},
		},
		{
			name: "a repeated name resolves to nothing",
			variables: []Template{
				{Name: "loki_datasource", Type: "datasource", Query: "loki"},
				{Name: "loki_datasource", Type: "datasource", Query: "loki"},
			},
			want: map[string]string{},
		},
		{
			name:      "no variables",
			variables: nil,
			want:      map[string]string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := dashboardFromVariables(t, tc.variables)
			if got := getDatasourceVariables(d); !equalStringMap(got, tc.want) {
				t.Errorf("getDatasourceVariables() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGetTemplateVariableName(t *testing.T) {
	for _, tc := range []struct {
		ref  string
		want string
	}{
		{"${loki_datasource}", "loki_datasource"},
		{"${loki_datasource:raw}", "loki_datasource"},
		{"$loki_datasource", "loki_datasource"},
		{"loki_datasource", "loki_datasource"},
		{"literal-uid", "literal-uid"},
		{"prefix-${name}", "prefix-${name}"},
		{"${unclosed", "${unclosed"},
		{"", ""},
	} {
		if got := getTemplateVariableName(tc.ref); got != tc.want {
			t.Errorf("getTemplateVariableName(%q) = %q, want %q", tc.ref, got, tc.want)
		}
	}
}

func TestTargetIsPrometheus(t *testing.T) {
	d := dashboardFromVariables(t, []Template{
		{Name: "prometheus_datasource", Type: "datasource", Query: "prometheus"},
		{Name: "loki_datasource", Type: "datasource", Query: "loki"},
	})

	for _, tc := range []struct {
		name   string
		target Target
		want   bool
	}{
		{
			name:   "the datasource variable decides",
			target: Target{Datasource: "${prometheus_datasource}"},
			want:   true,
		},
		{
			name:   "a loki variable is not prometheus",
			target: Target{Datasource: "${loki_datasource}"},
			want:   false,
		},
		{
			name:   "an unresolved target falls back to the dashboard datasource",
			target: Target{Datasource: "prometheus-uid"},
			want:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := d.targetIsPrometheus(tc.target); got != tc.want {
				t.Errorf("targetIsPrometheus() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTargetIsLoki(t *testing.T) {
	d := dashboardFromVariables(t, []Template{
		{Name: "prometheus_datasource", Type: "datasource", Query: "prometheus"},
		{Name: "loki_datasource", Type: "datasource", Query: "loki"},
	})

	for _, tc := range []struct {
		name   string
		target Target
		want   bool
	}{
		{
			name:   "a loki variable is loki",
			target: Target{Datasource: "${loki_datasource}"},
			want:   true,
		},
		{
			name:   "a prometheus variable is not loki",
			target: Target{Datasource: "${prometheus_datasource}"},
			want:   false,
		},
		{
			name:   "a datasource object with the loki type is loki",
			target: Target{Datasource: map[string]interface{}{"uid": "loki-uid", "type": "loki"}},
			want:   true,
		},
		{
			name:   "an unresolved target falls back to the dashboard datasource",
			target: Target{Datasource: "loki-uid"},
			want:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := d.targetIsLoki(tc.target); got != tc.want {
				t.Errorf("targetIsLoki(%v) = %v, want %v", tc.target.Datasource, got, tc.want)
			}
		})
	}
}

func equalStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
