package lint

import (
	"testing"
)

func TestPanelUnits(t *testing.T) {
	linter := NewPanelUnitsRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		{
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
	} {
		testRule(t, linter, Dashboard{Title: "test", Panels: []Panel{tc.panel}}, tc.result)
	}
}
