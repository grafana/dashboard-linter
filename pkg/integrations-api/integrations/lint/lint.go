package lint

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/cloud-onboarding/pkg/clients/grafana"
)

type Severity int

const (
	Success Severity = iota
	Exclude
	Warning
	Error
)

// Target is a deliberately incomplete representation of the Dashboard -> Panel -> Target type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Target struct {
	Idx  int    // This is the only (best?) way to uniquely identify a target, it is set by
	Expr string `json:"expr,omitempty"`
}

// Panel is a deliberately incomplete representation of the Dashboard -> Panel type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Panel struct {
	Title   string   `json:"title"`
	Targets []Target `json:"targets,omitempty"`
}

// Row is a deliberately incomplete representation of the Dashboard -> Row type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Row struct {
	Panels []Panel `json:"panels,omitempty"`
}

// Dashboard is a deliberately incomplete representation of the Dashboard type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Dashboard struct {
	Title  string  `json:"title,omitempty"`
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

func NewDashboardFromGrafanaDashboard(d grafana.GrafanaBoard) (Dashboard, error) {
	var dash Dashboard
	jsDash, err := json.Marshal(d.Dashboard)
	if err != nil {
		return dash, fmt.Errorf("unable to marshal dashboard back to json string: %w", err)
	}
	err = json.Unmarshal(jsDash, &dash)
	if err != nil {
		return dash, err
	}
	return dash, nil
}
