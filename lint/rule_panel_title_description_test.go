package lint

import (
	"testing"
)

func TestPanelTitleDescription(t *testing.T) {
	linter := NewPanelTitleDescriptionRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel with id '1' has missing title or description, currently has title '' and description: ''",
			},
			panel: Panel{
				Type:        "singlestat",
				Id:          1,
				Datasource:  "foo",
				Title:       "",
				Description: "",
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Type:        "singlestat",
				Id:          1,
				Datasource:  "foo",
				Title:       "testpanel",
				Description: "testdescription",
			},
		},
	} {
		testRule(t, linter, Dashboard{Title: "test", Panels: []Panel{tc.panel}}, tc.result)
	}
}
