package lint

import (
	"testing"
)

func TestPanelUnits(t *testing.T) {
	linter := NewPanelUnitsRule()

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
				FieldConfig: FieldConfig{
					Defaults: Defaults{
						Unit: "MyInvalidUnit",
					},
				},
			},
		},
		{
			name: "missing unit",
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
			name: "valid",
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
				FieldConfig: FieldConfig{
					Defaults: Defaults{
						Unit: "short",
					},
				},
			},
		},
		{
			name: "none - scalar",
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			testRule(t, linter, Dashboard{Title: "test", Panels: []Panel{tc.panel}}, tc.result)
		})
	}
}
