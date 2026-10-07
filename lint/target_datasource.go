package lint

import "strings"

// The datasource variables of a dashboard resolve the datasource reference
// of a target. A dashboard that queries more than one datasource carries one
// datasource variable per datasource, and each target names the datasource
// variable it queries through. The dashboard-level datasource template only
// decides for the targets whose reference does not resolve through a
// variable, which keeps the behavior of the single-datasource dashboards
// that the dashboard-level guards were written for.

// datasourceTemplateType is the template type of a datasource variable.
const datasourceTemplateType = "datasource"

// getDatasourceVariables returns the map from datasource variable name to the
// datasource plugin id. The query of a datasource template carries the plugin
// id. A dashboard that repeats a datasource variable name is malformed: the
// name resolves to nothing.
func getDatasourceVariables(d Dashboard) map[string]string {
	counts := map[string]int{}
	for _, t := range d.Templating.List {
		if t.Type != datasourceTemplateType || t.Name == "" || t.Query == "" {
			continue
		}
		counts[t.Name]++
	}

	vars := make(map[string]string, len(counts))
	for _, t := range d.Templating.List {
		if t.Type != datasourceTemplateType || t.Name == "" || t.Query == "" {
			continue
		}
		if counts[t.Name] > 1 {
			continue // a repeated name has no single plugin id
		}
		vars[t.Name] = t.Query
	}
	return vars
}

// getTemplateVariableName strips the Grafana variable syntax from a
// datasource reference. The forms are "${name}", "${name:format}", "$name",
// and the bare name. A malformed reference comes back unchanged and then
// misses the variable map.
func getTemplateVariableName(ref string) string {
	if name, ok := strings.CutPrefix(ref, "${"); ok {
		if !strings.HasSuffix(name, "}") {
			return ref
		}
		name = name[:len(name)-1]
		if base, _, found := strings.Cut(name, ":"); found {
			return base
		}
		return name
	}
	return strings.TrimPrefix(ref, "$")
}

// targetPlugin resolves the datasource reference of the target through the
// datasource variables of the dashboard and returns the datasource plugin
// id. The ok result is false when the target carries no datasource, an
// invalid datasource, or a reference that no datasource variable defines: a
// literal uid, an unknown variable, or a datasource without a uid.
func targetPlugin(d Dashboard, t Target) (string, bool) {
	ds, err := t.GetDataSource()
	if err != nil || ds.UID == "" {
		return "", false
	}
	plugin, ok := getDatasourceVariables(d)[getTemplateVariableName(ds.UID)]
	return plugin, ok
}

// targetIsPrometheus reports whether the target queries a Prometheus
// datasource. The datasource variable that the reference names decides, and
// the dashboard-level datasource template decides for the targets that do
// not resolve.
func (d Dashboard) targetIsPrometheus(t Target) bool {
	if plugin, ok := targetPlugin(d, t); ok {
		return plugin == Prometheus
	}
	template := getTemplateDatasource(d)
	return template != nil && template.Query == Prometheus
}

// targetIsLoki reports whether the target queries a Loki datasource. The
// datasource variable that the reference names decides. A target with a
// datasource object that carries the Loki type is a Loki target, and the
// dashboard-level datasource template decides for the remaining targets,
// which keeps the behavior of the Loki dashboards.
func (d Dashboard) targetIsLoki(t Target) bool {
	if plugin, ok := targetPlugin(d, t); ok {
		return plugin == Loki
	}
	if ds, err := t.GetDataSource(); err == nil && ds.Type == Loki {
		return true
	}
	template := getTemplateDatasource(d)
	return template != nil && template.Query == Loki
}
