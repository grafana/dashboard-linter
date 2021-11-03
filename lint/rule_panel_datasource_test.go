package lint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPanelDatasource(t *testing.T) {
	linter := NewPanelDatasourceRule()

	for _, tc := range []struct {
		result Result
		panel  Panel
	}{
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test', panel 'bar' does not use templates datasource, uses 'foo'",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "foo",
				Title:      "bar",
			},
		},
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			panel: Panel{
				Type:       "singlestat",
				Datasource: "$datasource",
			},
		},
	} {
		require.Equal(t, tc.result, linter.LintPanel(Dashboard{Title: "test"}, tc.panel))
	}
}
