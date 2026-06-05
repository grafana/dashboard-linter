package lint

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	dashv2 "github.com/grafana/grafana/apps/dashboard/pkg/apis/dashboard/v2"
)

// isV2APIVersion reports whether a kubernetes-shaped dashboard uses the v2
// schema (e.g. "dashboard.grafana.app/v2", "v2alpha1", "v2beta1"). The v2 spec
// is structurally different from the classic dashboard and is handled by a
// dedicated adapter.
func isV2APIVersion(apiVersion string) bool {
	v := apiVersion
	if i := strings.LastIndex(v, "/"); i >= 0 {
		v = v[i+1:]
	}
	return strings.HasPrefix(v, "v2")
}

// newDashboardFromV2 converts a v2 dashboard spec into the linter's internal
// Dashboard model so that all existing rules can run against it unchanged.
func newDashboardFromV2(spec json.RawMessage, apiVersion string) (Dashboard, error) {
	var s dashv2.DashboardSpec
	if err := json.Unmarshal(spec, &s); err != nil {
		return Dashboard{}, fmt.Errorf("parsing v2 dashboard spec: %w", err)
	}

	d := Dashboard{
		Title:      s.Title,
		APIVersion: apiVersion,
		Panels:     panelsFromV2(s.Elements),
	}
	if s.Editable != nil {
		d.Editable = *s.Editable
	}
	d.Templating.List = templatesFromV2(s.Variables)
	d.Annotations.List = annotationsFromV2(s.Annotations)
	return d, nil
}

// panelsFromV2 converts the v2 element map into the linter's panel slice.
// Library panels are skipped because they carry no inline spec to lint.
// Panels are sorted by id so output is deterministic (the element map has no
// inherent order; layout/order is irrelevant to linting).
func panelsFromV2(elements map[string]dashv2.DashboardElement) []Panel {
	var panels []Panel
	for _, el := range elements {
		if el.PanelKind == nil {
			continue
		}
		panels = append(panels, panelFromV2(el.PanelKind.Spec))
	}
	sort.Slice(panels, func(i, j int) bool { return panels[i].Id < panels[j].Id })
	return panels
}

func panelFromV2(ps dashv2.DashboardPanelSpec) Panel {
	p := Panel{
		Id:          int(ps.Id),
		Title:       ps.Title,
		Description: ps.Description,
		// In v2 the panel type is the visualization plugin id ("timeseries",
		// "stat", "table", "text", ...).
		Type:    ps.VizConfig.Group,
		Targets: targetsFromV2(ps.Data.Spec.Queries),
	}

	// The v2 fieldConfig and panel options mirror the classic JSON shape, so a
	// JSON round-trip is the most robust way to populate them.
	if fc, err := json.Marshal(ps.VizConfig.Spec.FieldConfig); err == nil {
		var lf FieldConfig
		if json.Unmarshal(fc, &lf) == nil {
			p.FieldConfig = &lf
		}
	}
	if opts, err := json.Marshal(ps.VizConfig.Spec.Options); err == nil {
		p.Options = opts
	}

	// Classic panels carry a panel-level datasource that the panel-datasource
	// rule inspects. v2 only has per-query datasources, so derive the panel
	// datasource from the first query.
	// TODO: add support for multiple datasources per panel in the rules and remove this hack.
	if len(p.Targets) > 0 {
		p.Datasource = p.Targets[0].Datasource
	}
	return p
}

func targetsFromV2(queries []dashv2.DashboardPanelQueryKind) []Target {
	var targets []Target
	for _, q := range queries {
		targets = append(targets, Target{
			RefId:      q.Spec.RefId,
			Hide:       q.Spec.Hidden,
			Expr:       stringFromQuerySpec(q.Spec.Query, "expr"),
			Datasource: datasourceFromV2(q.Spec.Query),
		})
	}
	return targets
}

// datasourceFromV2 maps a v2 DataQuery datasource into the shape the linter's
// GetDataSource understands: a map with "uid" (the templated reference, e.g.
// "$datasource") and "type" (the query group, e.g. "prometheus"/"loki").
// Returns nil when there is no datasource reference.
func datasourceFromV2(q dashv2.DashboardDataQueryKind) interface{} {
	if q.Datasource == nil || q.Datasource.Name == nil || *q.Datasource.Name == "" {
		return nil
	}
	m := map[string]interface{}{"uid": *q.Datasource.Name}
	if q.Group != "" {
		m["type"] = q.Group
	}
	return m
}

func templatesFromV2(vars []dashv2.DashboardVariableKind) []Template {
	var templates []Template
	for _, v := range vars {
		if t, ok := templateFromV2(v); ok {
			templates = append(templates, t)
		}
	}
	return templates
}

func templateFromV2(v dashv2.DashboardVariableKind) (Template, bool) {
	switch {
	case v.QueryVariableKind != nil:
		s := v.QueryVariableKind.Spec
		return Template{
			Type:       "query",
			Name:       s.Name,
			Label:      deref(s.Label),
			Multi:      s.Multi,
			Query:      stringFromQuerySpec(s.Query, "query"),
			Datasource: datasourceFromV2(s.Query),
		}, true
	case v.DatasourceVariableKind != nil:
		s := v.DatasourceVariableKind.Spec
		return Template{
			Type:  "datasource",
			Name:  s.Name,
			Label: deref(s.Label),
			Multi: s.Multi,
			// The datasource type drives prometheus/loki detection in the rules.
			Query: s.PluginId,
		}, true
	case v.CustomVariableKind != nil:
		s := v.CustomVariableKind.Spec
		return Template{Type: "custom", Name: s.Name, Label: deref(s.Label), Multi: s.Multi, Query: s.Query}, true
	case v.IntervalVariableKind != nil:
		s := v.IntervalVariableKind.Spec
		return Template{Type: "interval", Name: s.Name, Label: deref(s.Label), Query: s.Query}, true
	case v.ConstantVariableKind != nil:
		s := v.ConstantVariableKind.Spec
		return Template{Type: "constant", Name: s.Name, Label: deref(s.Label), Query: s.Query}, true
	case v.TextVariableKind != nil:
		s := v.TextVariableKind.Spec
		return Template{Type: "textbox", Name: s.Name, Label: deref(s.Label), Query: s.Query}, true
	case v.AdhocVariableKind != nil:
		s := v.AdhocVariableKind.Spec
		return Template{Type: "adhoc", Name: s.Name, Label: deref(s.Label)}, true
	case v.GroupByVariableKind != nil:
		s := v.GroupByVariableKind.Spec
		return Template{Type: "groupby", Name: s.Name, Label: deref(s.Label), Multi: s.Multi}, true
	case v.SwitchVariableKind != nil:
		s := v.SwitchVariableKind.Spec
		return Template{Type: "switch", Name: s.Name, Label: deref(s.Label)}, true
	}
	return Template{}, false
}

func annotationsFromV2(anns []dashv2.DashboardAnnotationQueryKind) []Annotation {
	var out []Annotation
	for _, a := range anns {
		out = append(out, Annotation{
			Name:       a.Spec.Name,
			Datasource: datasourceFromV2(a.Spec.Query),
		})
	}
	return out
}

// stringFromQuerySpec extracts a string field from a v2 DataQuery's free-form
// spec map (the datasource-specific payload, e.g. {"expr": "..."} for a panel
// query or {"query": "label_values(...)"} for a query variable).
func stringFromQuerySpec(q dashv2.DashboardDataQueryKind, key string) string {
	if q.Spec == nil {
		return ""
	}
	if s, ok := q.Spec[key].(string); ok {
		return s
	}
	return ""
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
