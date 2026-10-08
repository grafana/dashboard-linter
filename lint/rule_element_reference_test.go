package lint

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// v2ElementReferencePanel renders one defined panel of spec.elements, in the
// shape the v2 adapter parses.
func v2ElementReferencePanel(name string, id int) string {
	return fmt.Sprintf(`"%s": {
		"kind": "Panel",
		"spec": {
			"id": %d,
			"title": "%s",
			"data": { "kind": "QueryGroup", "spec": { "queries": [] } },
			"vizConfig": {
				"kind": "VizConfig", "group": "timeseries", "version": "1.0",
				"spec": { "options": {}, "fieldConfig": { "defaults": {}, "overrides": [] } }
			}
		}
	}`, name, id, name)
}

// v2ElementReferenceDashboard renders a kubernetes v2 dashboard with the
// named elements and the given layout JSON. The layout carries the
// ElementReference nodes that the rule walks.
func v2ElementReferenceDashboard(t *testing.T, elements []string, layout string) Dashboard {
	t.Helper()

	definitions := make([]string, 0, len(elements))
	for i, name := range elements {
		definitions = append(definitions, v2ElementReferencePanel(name, i+1))
	}

	raw := fmt.Sprintf(`{
		"apiVersion": "dashboard.grafana.app/v2beta1",
		"kind": "Dashboard",
		"spec": {
			"title": "element references",
			"elements": { %s },
			"layout": %s
		}
	}`, strings.Join(definitions, ", "), layout)

	d, err := NewDashboard([]byte(raw))
	require.NoError(t, err)
	return d
}

// lintElementReference runs the rule on the dashboard and returns the errors
// and the warnings of the results.
func lintElementReference(t *testing.T, d Dashboard) (errors, warnings []string) {
	t.Helper()

	rule := NewElementReferenceRule()
	rs := ResultSet{}
	rule.Lint(d, &rs)
	require.Len(t, rs.results, 1)

	for _, res := range rs.results[0].Result.Results {
		if res.Severity == Error {
			errors = append(errors, res.Message)
		}
		if res.Severity == Warning {
			warnings = append(warnings, res.Message)
		}
	}
	return errors, warnings
}

func TestElementReferenceRuleSkipsClassicDashboards(t *testing.T) {
	t.Parallel()

	d, err := NewDashboard([]byte(`{"title": "classic"}`))
	require.NoError(t, err)

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	assert.Empty(t, warnings)
}

func TestElementReferenceRuleSkipsV2WithoutSpec(t *testing.T) {
	t.Parallel()

	// A v2 dashboard whose spec could not be parsed reaches no rule except
	// v2-required-fields-rule; the empty-spec branch guards the direct use
	// of the rule.
	d := Dashboard{APIVersion: "dashboard.grafana.app/v2beta1"}

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	assert.Empty(t, warnings)
}

func TestElementReferenceRulePassesPlacedElements(t *testing.T) {
	t.Parallel()

	layout := `{"kind": "GridLayout", "spec": {"items": [
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "panel-1"}, "x": 0, "y": 0, "width": 12, "height": 8}},
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "panel-2"}, "x": 12, "y": 0, "width": 12, "height": 8}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1", "panel-2"}, layout)

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	assert.Empty(t, warnings)
}

func TestElementReferenceRulePassesNestedLayoutNodes(t *testing.T) {
	t.Parallel()

	// The walk visits the references wherever they sit in the layout, here
	// under a nested rows and tabs structure.
	layout := `{"kind": "RowsLayout", "spec": {"rows": [
		{"kind": "RowsLayoutRow", "spec": {"title": "row", "elements": [
			{"kind": "ElementReference", "name": "panel-1"}
		]}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1"}, layout)

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	assert.Empty(t, warnings)
}

func TestElementReferenceRuleReportsMissingElements(t *testing.T) {
	t.Parallel()

	layout := `{"kind": "GridLayout", "spec": {"items": [
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "panel-1"}, "x": 0, "y": 0, "width": 12, "height": 8}},
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "ghost"}, "x": 12, "y": 0, "width": 12, "height": 8}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1"}, layout)

	errors, warnings := lintElementReference(t, d)
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0], `layout references panel "ghost", which is not defined in spec.elements`)
	assert.Empty(t, warnings)
}

func TestElementReferenceRuleSortsAndDeduplicatesMissingElements(t *testing.T) {
	t.Parallel()

	// "ghost" is referenced twice and still reported once; the two missing
	// names come back sorted.
	layout := `{"kind": "GridLayout", "spec": {"items": [
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "panel-1"}, "x": 0, "y": 0, "width": 3, "height": 8}},
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "zombie"}, "x": 3, "y": 0, "width": 3, "height": 8}},
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "ghost"}, "x": 6, "y": 0, "width": 3, "height": 8}},
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "ghost"}, "x": 9, "y": 0, "width": 3, "height": 8}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1"}, layout)

	errors, warnings := lintElementReference(t, d)
	require.Len(t, errors, 2)
	assert.Contains(t, errors[0], `layout references panel "ghost", which is not defined in spec.elements`)
	assert.Contains(t, errors[1], `layout references panel "zombie", which is not defined in spec.elements`)
	assert.Empty(t, warnings)
}

func TestElementReferenceRuleReportsOrphanedElements(t *testing.T) {
	t.Parallel()

	layout := `{"kind": "GridLayout", "spec": {"items": [
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "panel-1"}, "x": 0, "y": 0, "width": 12, "height": 8}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1", "panel-2"}, layout)

	errors, warnings := lintElementReference(t, d)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], `panel "panel-2" is defined in spec.elements but is not placed in the layout`)
	assert.Empty(t, errors)
}

func TestElementReferenceRuleSortsOrphanedElements(t *testing.T) {
	t.Parallel()

	layout := `{"kind": "GridLayout", "spec": {"items": [
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "panel-1"}, "x": 0, "y": 0, "width": 12, "height": 8}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1", "zombie", "ghost"}, layout)

	errors, warnings := lintElementReference(t, d)
	require.Len(t, warnings, 2)
	assert.Contains(t, warnings[0], `panel "ghost" is defined in spec.elements but is not placed in the layout`)
	assert.Contains(t, warnings[1], `panel "zombie" is defined in spec.elements but is not placed in the layout`)
	assert.Empty(t, errors)
}

func TestElementReferenceRuleReportsMissingAndOrphanedTogether(t *testing.T) {
	t.Parallel()

	layout := `{"kind": "GridLayout", "spec": {"items": [
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "ghost"}, "x": 0, "y": 0, "width": 12, "height": 8}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1"}, layout)

	errors, warnings := lintElementReference(t, d)
	require.Len(t, errors, 1)
	require.Len(t, warnings, 1)
	assert.Contains(t, errors[0], `"ghost"`)
	assert.Contains(t, warnings[0], `"panel-1"`)
}

func TestElementReferenceRuleIgnoresReferencesWithoutName(t *testing.T) {
	t.Parallel()

	// A nameless ElementReference records nothing, so the defined panel is
	// orphaned and nothing is missing.
	layout := `{"kind": "GridLayout", "spec": {"items": [
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": ""}, "x": 0, "y": 0, "width": 12, "height": 8}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1"}, layout)

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], `panel "panel-1" is defined in spec.elements but is not placed in the layout`)
}

func TestElementReferenceRuleIgnoresNonReferenceNodes(t *testing.T) {
	t.Parallel()

	// The panel definitions carry kind Panel and a name of their own; only
	// kind ElementReference counts as a reference.
	layout := `{"kind": "GridLayout", "spec": {"items": [
		{"kind": "GridLayoutItem", "spec": {"element": {"kind": "ElementReference", "name": "panel-1"}, "x": 0, "y": 0, "width": 12, "height": 8}}
	]}}`

	d := v2ElementReferenceDashboard(t, []string{"panel-1"}, layout)

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	assert.Empty(t, warnings)
}

func TestElementReferenceRulePassesDashboardWithoutLayout(t *testing.T) {
	t.Parallel()

	// A dashboard whose spec carries no layout references nothing: every
	// element is orphaned.
	d := v2ElementReferenceDashboard(t, []string{"panel-1", "panel-2"}, `{"kind": "GridLayout", "spec": {"items": []}}`)

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	require.Len(t, warnings, 2)
}

func TestElementReferenceRuleSkipsSpecWithoutElements(t *testing.T) {
	t.Parallel()

	// A spec with a layout but no elements key: nothing is defined, nothing
	// referenced, nothing reported.
	layout := `{"kind": "GridLayout", "spec": {"items": []}}`
	raw := fmt.Sprintf(`{
		"apiVersion": "dashboard.grafana.app/v2beta1",
		"kind": "Dashboard",
		"spec": {"title": "no elements", "layout": %s}
	}`, layout)

	d, err := NewDashboard([]byte(raw))
	require.NoError(t, err)

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	assert.Empty(t, warnings)
}

func TestElementReferenceRuleSkipsUnparseableSpec(t *testing.T) {
	t.Parallel()

	// The rule set skips the dashboards whose v2 spec could not be parsed,
	// so the rule defends itself against a spec that is not valid JSON when
	// it is called directly.
	d := Dashboard{
		APIVersion: "dashboard.grafana.app/v2beta1",
		Spec:       json.RawMessage(`{"elements": `),
	}

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	assert.Empty(t, warnings)
}

func TestElementReferenceRuleSkipsNonObjectSpec(t *testing.T) {
	t.Parallel()

	d := Dashboard{
		APIVersion: "dashboard.grafana.app/v2beta1",
		Spec:       json.RawMessage(`[1, 2]`),
	}

	errors, warnings := lintElementReference(t, d)
	assert.Empty(t, errors)
	assert.Empty(t, warnings)
}
