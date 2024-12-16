package lint

import (
	"testing"
)

func TestPanelUnits(t *testing.T) {
	linter := NewPanelUnitsRule()
	var overrides = make([]Override, 0)
	overrides = append(overrides, Override{
		OverrideProperties: []OverrideProperty{
			{
				Id: "mappings",
				Value: []byte(`[
						{
						"type": "value",
						"options": {
						"1": {
							"text": "OK",
							"color": "green",
							"index": 0
						},
						"2": {
							"text": "Problem",
							"color": "red",
							"index": 1
						}
						}
					}
				]`),
			},
		},
	})
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
					Defaults: Defaults{
						Unit: "MyInvalidUnit",
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
					Defaults: Defaults{
						Unit: "short",
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
					Defaults: Defaults{
						Unit: "none",
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
					Defaults: Defaults{
						Mappings: []byte(`
							[
								{
								"options": {
									"0": {
									"color": "red",
									"index": 1,
									"text": "DOWN"
									},
									"1": {
									"color": "green",
									"index": 0,
									"text": "UP"
									}
								},
								"type": "value"
								}
							]`,
						),
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
					Overrides: overrides,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testRule(t, linter, Dashboard{Title: "test", Panels: []Panel{tc.panel}}, tc.result)
		})
	}
}
