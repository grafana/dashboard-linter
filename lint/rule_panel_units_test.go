package lint

import (
	"testing"

	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
)

func ptr[T any](t T) *T { return &t }
func TestPanelUnits(t *testing.T) {
	linter := NewPanelUnitsRule()

	testValueMap := &dashboard.ValueMap{
		Type: "value",
		Options: map[string]dashboard.ValueMappingResult{
			"1": {
				Text:  ptr("Ok"),
				Color: ptr("green"),
			},
			"2": {
				Text:  ptr("Down"),
				Color: ptr("red"),
			},
		},
	}

	for _, tc := range []struct {
		name   string
		result Result
		panel  Panel
	}{
		{
			name: "invalid unit",
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' has no or invalid units defined: 'MyInvalidUnit'",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
				FieldConfig: &FieldConfig{
					Defaults: dashboard.FieldConfig{
						Unit: ptr("MyInvalidUnit"),
					},
				},
			},
		},
		{
			name: "missing FieldConfig",
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' has no or invalid units defined: ''",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
			},
		},
		{
			name: "empty FieldConfig",
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' has no or invalid units defined: ''",
			},
			panel: Panel{
				Type:        "singlestat",
				Datasource:  "foo",
				Title:       "bar",
				FieldConfig: &FieldConfig{},
			},
		},
		{
			name:   "valid",
			result: ResultSuccess,
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
				FieldConfig: &FieldConfig{
					Defaults: dashboard.FieldConfig{
						Unit: ptr("short"),
					},
				},
			},
		},
		{
			name:   "none - scalar",
			result: ResultSuccess,
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
				FieldConfig: &FieldConfig{
					Defaults: dashboard.FieldConfig{
						Unit: ptr("none"),
					},
				},
			},
		},
		{
			name:   "has nonnumeric reduceOptions fields",
			result: ResultSuccess,
			panel: Panel{
				Type:       "stat",
				Datasource: "foo",
				Title:      "bar",
				Options: []byte(`
					{
						"reduceOptions": {
							"fields": "/^version$/"
						}
					}
					
				`),
			},
		},
		{
			name: "has empty reduceOptions fields(Numeric Fields default value)",
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' has no or invalid units defined: ''",
			},
			panel: Panel{
				Type:       "stat",
				Datasource: "foo",
				Title:      "bar",
				Options: []byte(`
					{
						"reduceOptions": {
							"fields": ""
						}
					}
					
				`),
			},
		},
		{
			name:   "no units but have value mappings",
			result: ResultSuccess,
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
				FieldConfig: &FieldConfig{
					Defaults: dashboard.FieldConfig{
						Mappings: []dashboard.ValueMapOrRangeMapOrRegexMapOrSpecialValueMap{
							dashboard.ValueMapOrRangeMapOrRegexMapOrSpecialValueMap{
								ValueMap: testValueMap,
							},
						},
					},
				},
			},
		},
		{
			name:   "no units but have value mappings in overrides",
			result: ResultSuccess,
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
				FieldConfig: &FieldConfig{
					Overrides: []dashboard.DashboardFieldConfigSourceOverrides{
						dashboard.DashboardFieldConfigSourceOverrides{
							Matcher: dashboard.MatcherConfig{
								Id:      "byRegexp",
								Options: "/.*/",
							},
							Properties: []dashboard.DynamicConfigValue{
								dashboard.DynamicConfigValue{
									Id: "mappings",
									Value: []dashboard.ValueMapOrRangeMapOrRegexMapOrSpecialValueMap{
										dashboard.ValueMapOrRangeMapOrRegexMapOrSpecialValueMap{
											ValueMap: testValueMap,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testRule(t, linter, Dashboard{Title: "test", Panels: []Panel{tc.panel}}, tc.result)
		})
	}
}
