package lint

import (
	"encoding/json"
	"fmt"
	"sort"
)

// NewElementReferenceRule checks that the dashboard layout only references
// panels that are defined in the dashboard spec, and that every defined panel
// is placed in the layout. The v2 CUE schema does not check these references,
// so a stale reference only shows up as a broken dashboard at render time.
func NewElementReferenceRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "element-reference-rule",
		description: "Checks that layout element references resolve to defined panels, and that every panel is placed in the layout.",
		fn: func(d Dashboard) DashboardRuleResults {
			r := DashboardRuleResults{}

			if !isV2APIVersion(d.APIVersion) || len(d.Spec) == 0 {
				return r
			}

			elements, tree, ok := decodeSpec(d.Spec)
			if !ok {
				return r
			}

			referenced := collectElementReferences(tree)

			for _, name := range missingElements(referenced, elements) {
				r.AddError(d, fmt.Sprintf("layout references panel %q, which is not defined in spec.elements", name))
			}

			for _, name := range orphanedElements(referenced, elements) {
				r.AddWarning(d, fmt.Sprintf("panel %q is defined in spec.elements but is not placed in the layout", name))
			}

			return r
		},
	}
}

// decodeSpec decodes the raw dashboard spec. It returns the names of the
// defined elements and the decoded document as a generic tree, so the
// reference walk visits the layout wherever it sits in the spec. The ok
// result is false when the spec is not a JSON object.
func decodeSpec(spec json.RawMessage) (elements map[string]any, tree any, ok bool) {
	var doc any
	if err := json.Unmarshal(spec, &doc); err != nil {
		return nil, nil, false
	}

	root, ok := doc.(map[string]any)
	if !ok {
		// A JSON object is the only useful shape of a dashboard spec; the
		// v2-required-fields-rule reports the other shapes.
		return nil, nil, false
	}

	elements, _ = root["elements"].(map[string]any)
	return elements, root, true
}

// collectElementReferences returns the set of element names referenced by
// ElementReference nodes anywhere in the decoded dashboard spec.
func collectElementReferences(tree any) map[string]bool {
	referenced := map[string]bool{}
	walkElementReferences(referenced, tree)
	return referenced
}

// walkElementReferences visits node and each child depth first.
func walkElementReferences(referenced map[string]bool, node any) {
	switch typed := node.(type) {
	case map[string]any:
		addElementReference(referenced, typed)
		for _, child := range typed {
			walkElementReferences(referenced, child)
		}
	case []any:
		for _, child := range typed {
			walkElementReferences(referenced, child)
		}
	}
}

// addElementReference records the name of an ElementReference node, if node
// is one. A reference without a name records nothing.
func addElementReference(referenced map[string]bool, node map[string]any) {
	if kind, _ := node["kind"].(string); kind != "ElementReference" {
		return
	}

	if name, _ := node["name"].(string); name != "" {
		referenced[name] = true
	}
}

// missingElements returns the sorted names of referenced elements that have
// no definition in elements.
func missingElements(referenced map[string]bool, elements map[string]any) []string {
	missing := make([]string, 0, len(referenced))
	for name := range referenced {
		if _, ok := elements[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)

	return missing
}

// orphanedElements returns the sorted names of defined elements that no
// layout node references.
func orphanedElements(referenced map[string]bool, elements map[string]any) []string {
	orphaned := make([]string, 0, len(elements))
	for name := range elements {
		if !referenced[name] {
			orphaned = append(orphaned, name)
		}
	}
	sort.Strings(orphaned)

	return orphaned
}
