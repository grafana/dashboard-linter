package lint

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Severity int

const (
	Success Severity = iota
	Exclude
	Quiet
	Warning
	Error
	Fixed

	Prometheus = "prometheus"
	Loki       = "loki"
)

// Target is a deliberately incomplete representation of the Dashboard -> Template type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Template struct {
	Name       string             `json:"name"`
	Label      string             `json:"label"`
	Type       string             `json:"type"`
	RawQuery   interface{}        `json:"query"`
	Query      string             `json:"-"`
	Datasource interface{}        `json:"datasource,omitempty"`
	Multi      bool               `json:"multi"`
	AllValue   string             `json:"allValue,omitempty"`
	Current    RawTemplateValue   `json:"current"`
	Options    []RawTemplateValue `json:"options"`
	Refresh    int                `json:"refresh"`
	// If you add properties here don't forget to add them to the raw struct, and assign them from raw to actual in UnmarshalJSON below!
}

type RawTemplateValue map[string]interface{}

type TemplateValue struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

func (t *Template) UnmarshalJSON(buf []byte) error {
	var raw struct {
		Name       string             `json:"name"`
		Label      string             `json:"label"`
		Type       string             `json:"type"`
		Query      interface{}        `json:"query"`
		Datasource interface{}        `json:"datasource,omitempty"`
		Multi      bool               `json:"multi"`
		AllValue   string             `json:"allValue"`
		Current    RawTemplateValue   `json:"current"`
		Options    []RawTemplateValue `json:"options"`
		Refresh    int                `json:"refresh"`
	}
	if err := json.Unmarshal(buf, &raw); err != nil {
		return err
	}

	t.Name = raw.Name
	t.Label = raw.Label
	t.Type = raw.Type
	t.Datasource = raw.Datasource
	t.Multi = raw.Multi
	t.AllValue = raw.AllValue
	t.Current = raw.Current
	t.Options = raw.Options
	t.Refresh = raw.Refresh
	t.RawQuery = raw.Query

	// the 'adhoc' and 'custom' variable type does not have a field `Query`, so we can't perform these checks
	if t.Type != "adhoc" && t.Type != "custom" {
		switch v := raw.Query.(type) {
		case string:
			t.Query = v
		case map[string]interface{}:
			query, ok := v[targetTypeQuery]
			if ok {
				t.Query = query.(string)
			}
		default:
			return fmt.Errorf("invalid type for field 'query': %v", v)
		}
	}

	return nil
}

func (t *Template) GetDataSource() (Datasource, error) {
	return GetDataSource(t.Datasource)
}

func (raw *RawTemplateValue) Get() (TemplateValue, error) {
	t := TemplateValue{}
	var txt, val interface{}
	m := *raw

	txt, ok := m["text"]
	if ok {
		switch tt := txt.(type) {
		case string:
			t.Text = txt.(string)
		case []interface{}:
			t.Text = txt.([]interface{})[0].(string)
		default:
			return t, fmt.Errorf("invalid type for field 'text': %v", tt)
		}
	}

	val, ok = m["value"]
	if ok {
		switch vt := val.(type) {
		case string:
			t.Value = val.(string)
		case []interface{}:
			t.Value = val.([]interface{})[0].(string)
		default:
			return t, fmt.Errorf("invalid type for field 'value': %v", vt)
		}
	}

	return t, nil
}

// Input is a deliberately incomplete representation of the Dashboard -> Input type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Input struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	PluginID string `json:"pluginId"`
}

type Datasource struct {
	UID  string `json:"uid"`
	Type string `json:"type"`
}

func GetDataSource(raw interface{}) (Datasource, error) {
	switch v := raw.(type) {
	case nil:
		return Datasource{}, nil
	case string:
		return Datasource{v, ""}, nil
	case map[string]interface{}:
		uid, ok := v["uid"]
		if !ok {
			return Datasource{}, fmt.Errorf("invalid type for field 'datasource': missing uid field")
		}
		uidStr, ok := uid.(string)
		if !ok {
			return Datasource{}, fmt.Errorf("invalid type for field 'datasource': invalid uid field type, should be string")
		}
		if dsType, ok := v["type"]; ok {
			dsTypeStr, ok := dsType.(string)
			if !ok {
				return Datasource{}, fmt.Errorf("invalid type for field 'datasource': invalid type field type, should be string")
			}
			return Datasource{uidStr, dsTypeStr}, nil
		}
		return Datasource{uidStr, ""}, nil
	default:
		return Datasource{}, fmt.Errorf("invalid type for field 'datasource': %v", v)
	}
}

// Target is a deliberately incomplete representation of the Dashboard -> Panel -> Target type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Target struct {
	Idx        int         `json:"-"` // This is the only (best?) way to uniquely identify a target, it is set by GetPanels
	Datasource interface{} `json:"datasource,omitempty"`
	Expr       string      `json:"expr,omitempty"`
	PanelId    int         `json:"panelId,omitempty"`
	RefId      string      `json:"refId,omitempty"`
	Hide       bool        `json:"hide"`
}

func (t *Target) GetDataSource() (Datasource, error) {
	return GetDataSource(t.Datasource)
}

type Annotation struct {
	Name       string      `json:"name"`
	Datasource interface{} `json:"datasource,omitempty"`
}

func (a *Annotation) GetDataSource() (Datasource, error) {
	return GetDataSource(a.Datasource)
}

// Panel is a deliberately incomplete representation of the Dashboard -> Panel type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Panel struct {
	Id          int             `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Targets     []Target        `json:"targets,omitempty"`
	Datasource  interface{}     `json:"datasource,omitempty"`
	Type        string          `json:"type"`
	Panels      []Panel         `json:"panels,omitempty"`
	FieldConfig *FieldConfig    `json:"fieldConfig,omitempty"`
	Options     json.RawMessage `json:"options,omitempty"`
}

type FieldConfig struct {
	Defaults  Defaults   `json:"defaults,omitempty"`
	Overrides []Override `json:"overrides,omitempty"`
}

type Override struct {
	OverrideProperties []OverrideProperty `json:"properties"`
}

type OverrideProperty struct {
	Id    string `json:"id"`
	Value any    `json:"value"`
}

// oversimplified Reduce options
type ReduceOptions struct {
	Fields string   `json:"fields,omitempty"`
	Calcs  []string `json:"[]calcs,omitempty"`
	Values bool     `json:"values,omitempty"`
	Limit  int      `json:"limit,omitempty"`
}

// Stat panel options is a deliberately incomplete representation of the stat panel options from grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type StatOptions struct {
	ReduceOptions ReduceOptions `json:"reduceOptions,omitempty"`
}

type Defaults struct {
	Unit     string          `json:"unit,omitempty"`
	Mappings json.RawMessage `json:"mappings,omitempty"`
}

// GetPanels returns the all panels nested inside the panel (inc the current panel)
func (p *Panel) GetPanels() []Panel {
	panels := []Panel{*p}
	for _, panel := range p.Panels {
		panels = append(panels, panel.GetPanels()...)
	}
	return panels
}

func (p *Panel) GetDataSource() (Datasource, error) {
	return GetDataSource(p.Datasource)
}

// Row is a deliberately incomplete representation of the Dashboard -> Row type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Row struct {
	Panels []Panel `json:"panels,omitempty"`
}

// GetPanels returns the all panels nested inside the row
func (r *Row) GetPanels() []Panel {
	var panels []Panel
	for _, panel := range r.Panels {
		panels = append(panels, panel.GetPanels()...)
	}
	return panels
}

// Dashboard is a deliberately incomplete representation of the Dashboard type in grafana.
// The properties which are extracted from JSON are only those used for linting purposes.
type Dashboard struct {
	Inputs     []Input `json:"__inputs"`
	Title      string  `json:"title,omitempty"`
	Templating struct {
		List []Template `json:"list"`
	} `json:"templating"`
	Annotations struct {
		List []Annotation `json:"list"`
	} `json:"annotations"`
	Rows     []Row   `json:"rows,omitempty"`
	Panels   []Panel `json:"panels,omitempty"`
	Editable bool    `json:"editable,omitempty"`

	// Kubernetes shaped dashboards will include an APIVersion and Kind
	APIVersion string `json:"apiVersion,omitempty"`
	// When reading a kubernetes encoded dashboard, the Dashboard will be
	Spec json.RawMessage `json:"spec,omitempty"`
}

// GetPanels returns the all panels whether they are nested in the (now deprecated) "rows" property or
// in the top level "panels" property. This also monkeypatches Target.Idx into each panel which is used
// to uniquely identify panel targets while linting.
func (d *Dashboard) GetPanels() []Panel {
	var p []Panel
	for _, row := range d.Rows {
		p = append(p, row.GetPanels()...)
	}
	for _, panel := range d.Panels {
		p = append(p, panel.GetPanels()...)
	}
	for pi, pa := range p {
		for ti := range pa.Targets {
			p[pi].Targets[ti].Idx = ti
		}
	}
	return p
}

// GetTemplateByType returns all dashboard templates which match the provided type. Type comparison
// is case insensitive as it uses strings.EqualFold()
func (d *Dashboard) GetTemplateByType(t string) []Template {
	var retval []Template
	for _, templ := range d.Templating.List {
		if strings.EqualFold(templ.Type, t) {
			retval = append(retval, templ)
		}
	}
	return retval
}

func (d *Dashboard) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

func NewDashboard(buf []byte) (Dashboard, error) {
	var dash Dashboard
	if err := json.Unmarshal(buf, &dash); err != nil {
		return dash, err
	}
	// Support kubernetes flavored dashboards
	if dash.Spec != nil {
		apiVersion := dash.APIVersion
		if apiVersion != "" {
			if !(strings.HasPrefix(apiVersion, "v0") || strings.HasPrefix(apiVersion, "v1")) {
				return dash, fmt.Errorf("unsupported apiVersion")
			}
		}
		if err := json.Unmarshal(dash.Spec, &dash); err != nil {
			return dash, err
		}
		dash.APIVersion = apiVersion // preserve the original APIVersion
	}
	return dash, nil
}
