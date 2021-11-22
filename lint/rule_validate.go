package lint

import (
	"sync"

	"cuelang.org/go/cue/errors"
	"github.com/grafana/grafana/pkg/schema"
	"github.com/grafana/grafana/pkg/schema/load"
)

// Performance-guarding logic for loading the Grafana dashboard scuemata.
//
// We use these because the CUE runtime currently (as of v0.4.0) has performance issues with
// large disjunctions, which are a necessary part of how the "dist" dashboard schema composes
// in schema from Grafana plugins. These issues only manifest on initial load, so we a) do not
// do the load in an init() function, and b) guard it with a sync.Once in order to ensure that
// the cost is paid only once even for multiple calls.
var baseOnce, distOnce sync.Once
var basesch, distsch schema.CueSchema

func getBaseSchema() (schema.CueSchema, error) {
	var err error
	baseOnce.Do(func() {
		basesch, err = load.BaseDashboardFamily(load.GetDefaultLoadPaths())
		if err != nil {
			panic(err)
		}
	})
	return basesch, err
}

func getDistSchema() (schema.CueSchema, error) {
	var err error
	distOnce.Do(func() {
		distsch, err = load.DistDashboardFamily(load.GetDefaultLoadPaths())
		if err != nil {
			panic(err)
		}
	})
	return distsch, err
}

func NewDashboardValidateRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "validate-base",
		description: "validate-base checks that the entire dashboard JSON is valid according to the canonical base Grafana dashboard schema.",
		fn: func(d Dashboard) Result {
			// TODO this is more properly SearchAndValidate, but this works for now as
			// 1) there's only a single schema version so far and 2) the API is going to
			// change when it moves from github.com/grafana/grafana/pkg/schema
			// to github.com/grafana/scuemata.
			sch, err := getBaseSchema()
			if err != nil {
				panic(err)
			}
			err = sch.Validate(schema.Resource{
				Value: d.Raw,
			})

			if err != nil {
				return Result{
					Severity: Error,
					Message:  errors.Details(err, nil),
				}
			}

			return Result{
				Severity: Success,
				Message:  "OK",
			}
		},
	}
}

func NewDashboardValidateDistRule() *DashboardRuleFunc {
	return &DashboardRuleFunc{
		name:        "validate-dist",
		description: "validate-dist checks that the entire dashboard JSON is valid according to the canonical base Grafana dashboard schema and core plugin schemas.",
		fn: func(d Dashboard) Result {
			// TODO this is more properly SearchAndValidate, but this works for now as
			// 1) there's only a single schema version so far and 2) the API is going to
			// change when it moves from github.com/grafana/grafana/pkg/schema
			// to github.com/grafana/scuemata.
			sch, err := getDistSchema()
			if err != nil {
				panic(err)
			}
			err = sch.Validate(schema.Resource{
				Value: d.Raw,
			})

			if err != nil {
				return Result{
					Severity: Error,
					Message:  errors.Details(err, nil),
				}
			}

			return Result{
				Severity: Success,
				Message:  "OK",
			}
		},
	}
}
