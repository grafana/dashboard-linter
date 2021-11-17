package lint

import (
	"encoding/json"
	"fmt"
)

type Severity int

const (
	Success Severity = iota
	Exclude
	Warning
	Error
	Quiet
)

// Target is a deliberately incomplete representation of the Dashboard -> Template type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Template struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Type       string `json:"type"`
	Query      string `json:"query"`
	Datasource string `json:"datasource"`
	Multi      bool   `json:"multi"`
	AllValue   string `json:"allValue"`
}

func (t *Template) UnmarshalJSON(buf []byte) error {
	var raw struct {
		Name       string      `json:"name"`
		Label      string      `json:"label"`
		Type       string      `json:"type"`
		Query      interface{} `json:"query"`
		Datasource interface{} `json:"datasource"`
		Multi      bool        `json:"multi"`
		AllValue   string      `json:"allValue"`
	}

	if err := json.Unmarshal(buf, &raw); err != nil {
		return err
	}

	t.Name = raw.Name
	t.Label = raw.Label
	t.Type = raw.Type
	t.Multi = raw.Multi
	t.AllValue = raw.AllValue

	switch v := raw.Datasource.(type) {
	case nil:
		t.Datasource = ""
	case string:
		t.Datasource = v
	case map[string]interface{}:
		t.Datasource = v["uid"].(string)
	default:
		return fmt.Errorf("invalid type for field 'datasource': %v", v)
	}

	switch v := raw.Query.(type) {
	case string:
		t.Query = v
	case map[string]interface{}:
		t.Query = v["query"].(string)
	default:
		return fmt.Errorf("invalid type for field 'query': %v", v)
	}

	return nil
}

// Target is a deliberately incomplete representation of the Dashboard -> Panel -> Target type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Target struct {
	Idx  int    // This is the only (best?) way to uniquely identify a target, it is set by
	Expr string `json:"expr,omitempty"`
}

// Panel is a deliberately incomplete representation of the Dashboard -> Panel type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Panel struct {
	Title      string   `json:"title"`
	Targets    []Target `json:"targets,omitempty"`
	Datasource string   `json:"datasource"`
	Type       string   `json:"type"`
}

// Row is a deliberately incomplete representation of the Dashboard -> Row type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Row struct {
	Panels []Panel `json:"panels,omitempty"`
}

// Dashboard is a deliberately incomplete representation of the Dashboard type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Dashboard struct {
	Title      string `json:"title,omitempty"`
	Templating struct {
		List []Template `json:"list"`
	} `json:"templating"`
	Rows   []Row   `json:"rows,omitempty"`
	Panels []Panel `json:"panels,omitempty"`
}

// GetPanels returns the all panels whether they are nested in the (now deprecated) "rows" property or
// in the top level "panels" property. This also monkeypatches Target.Idx into each panel which is used
// to uniquely identify panel targets while linting.
func (d *Dashboard) GetPanels() []Panel {
	var p []Panel
	for _, row := range d.Rows {
		p = append(p, row.Panels...)
	}
	p = append(p, d.Panels...)
	for pi, pa := range p {
		for ti := range pa.Targets {
			p[pi].Targets[ti].Idx = ti
		}
	}
	return p
}

func NewDashboard(buf []byte) (Dashboard, error) {
	var dash Dashboard
	if err := json.Unmarshal(buf, &dash); err != nil {
		return dash, err
	}
	return dash, nil
}
