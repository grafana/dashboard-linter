package lint

import (
	"encoding/json"
	"fmt"
)

// v2RequiredField describes a required top-level DashboardSpec field and the
// JSON kind it must have (first byte of the raw value).
type v2RequiredField struct {
	name      string
	firstByte byte // '"' string, '[' array, '{' object, 't'/'f' bool
}

// v2RequiredFields lists every top-level DashboardSpec field that has no
// omitempty tag in the generated struct and is therefore required by the API.
var v2RequiredFields = []v2RequiredField{
	{"annotations", '['},
	{"cursorSync", '"'},
	{"elements", '{'},
	{"layout", '{'},
	{"links", '['},
	{"preload", 't'}, // bool — first byte is 't' (true) or 'f' (false)
	{"tags", '['},
	{"timeSettings", '{'},
	{"title", '"'},
	{"variables", '['},
}

// NewV2RequiredFieldsRule returns a rule that checks a v2 dashboard spec for
// missing or wrongly-typed required top-level fields. It mirrors the validation
// Grafana performs when the "Edit as code" dialog is saved.
func NewV2RequiredFieldsRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "v2-required-fields-rule",
		description: "Checks that v2 dashboards include all required spec fields.",
		fn: func(d Dashboard) DashboardRuleResults {
			r := DashboardRuleResults{}
			if !isV2APIVersion(d.APIVersion) || len(d.Spec) == 0 {
				return r
			}

			var raw map[string]json.RawMessage
			if err := json.Unmarshal(d.Spec, &raw); err != nil {
				r.AddError(d, "v2 spec must be a JSON object")
				return r
			}

			for _, f := range v2RequiredFields {
				val, ok := raw[f.name]
				if !ok {
					r.AddError(d, fmt.Sprintf("v2 spec is missing required property %q", f.name))
					continue
				}
				if len(val) == 0 {
					continue
				}
				first := val[0]
				// booleans can start with 't' or 'f'
				if f.firstByte == 't' {
					if first != 't' && first != 'f' {
						r.AddError(d, fmt.Sprintf("v2 spec property %q must be a boolean", f.name))
					}
					continue
				}
				if first != f.firstByte {
					r.AddError(d, fmt.Sprintf("v2 spec property %q has wrong type (expected %s)", f.name, jsonKindName(f.firstByte)))
				}
			}
			// If the spec had a parse error but none of the required-field checks
			// fired (e.g. a non-required field has the wrong type), surface the raw
			// parse error so the problem is still reported.
			if d.V2ParseError && len(r.Results) == 0 {
				r.AddError(d, fmt.Sprintf("v2 spec could not be parsed: %s", d.V2ParseErrorMsg))
			}
			return r
		},
	}
}

func jsonKindName(firstByte byte) string {
	switch firstByte {
	case '"':
		return "string"
	case '[':
		return "array"
	case '{':
		return "object"
	case 't', 'f':
		return "boolean"
	default:
		return "unknown"
	}
}
