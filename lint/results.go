package lint

import (
	"fmt"
	"sort"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

type Result struct {
	Severity Severity
	Message  string
}

// ResultContext is used by ResultSet to keep all the state data about a lint execution and it's results.
type ResultContext struct {
	Result      Result
	Rule        Rule
	Integration *integrations.Integration
	Dashboard   *Dashboard
	Panel       *Panel
	Target      *Target
}

func (r ResultContext) TtyPrint() {
	var sym string
	switch s := r.Result.Severity; s {
	case Success:
		sym = "✔️"
	case Exclude:
		sym = "➖"
	case Warning:
		sym = "⚠️"
	case Error:
		sym = "❌"
	case Quiet:
		return
	}

	fmt.Printf("[%s] Integration: %s - %s\n", sym, r.Integration.Meta.Slug, r.Result.Message)
}

type ResultSet struct {
	results []ResultContext
	config  *Configuration
}

func (rs *ResultSet) Configure(c *Configuration) {
	rs.config = c
}

func (rs *ResultSet) AddResult(r ResultContext) {
	rs.results = append(rs.results, r)
}

func (rs *ResultSet) MaximumSeverity() Severity {
	retVal := Success
	for _, res := range rs.results {
		if res.Result.Severity > retVal {
			retVal = res.Result.Severity
		}
	}
	return retVal
}

func (rs *ResultSet) ByRule() map[string][]ResultContext {
	ret := make(map[string][]ResultContext)
	for _, res := range rs.results {
		ret[res.Rule.Name()] = append(ret[res.Rule.Name()], res)
	}
	for _, rule := range ret {
		sort.SliceStable(rule, func(i, j int) bool {
			return rule[i].Integration.Meta.Slug < rule[j].Integration.Meta.Slug
		})
	}
	return ret
}

func (rs *ResultSet) ByIntegration() map[string][]ResultContext {
	ret := make(map[string][]ResultContext)
	for _, res := range rs.results {
		ret[res.Integration.Meta.Slug] = append(ret[res.Integration.Meta.Slug], res)
	}
	return ret
}

func (rs *ResultSet) ReportByRule() {
	for _, res := range rs.ByRule() {
		fmt.Println(res[0].Rule.Description())
		for _, r := range res {
			rs.config.Apply(r).TtyPrint()
		}
	}
}

func (rs *ResultSet) ReportByIntegration() {
	byIntegration := rs.ByIntegration()
	keys := make([]string, 0, len(byIntegration))
	for k := range byIntegration {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, slug := range keys {
		fmt.Printf("Integration: %s\n", slug)
		res := byIntegration[slug]
		for _, r := range res {
			fmt.Printf("  %s\n", r.Rule.Description())
			fmt.Print("    ")
			rs.config.Apply(r).TtyPrint()
		}
	}
}
