package lint

import (
	"fmt"
	"os"
	"sort"
)

var ResultSuccess = Result{
	Severity: Success,
	Message:  "OK",
}

type Result struct {
	Severity Severity
	Message  string
}

// ResultContext is used by ResultSet to keep all the state data about a lint execution and it's results.
type ResultContext struct {
	Result    Result
	Rule      Rule
	Dashboard *Dashboard
	Panel     *Panel
	Target    *Target
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

	fmt.Fprintf(os.Stdout, "[%s] %s\n", sym, r.Result.Message)
}

type ResultSet struct {
	results []ResultContext
	config  *ConfigurationFile
}

// Configure adds, and applies the provided configuration to all results currently in the ResultSet
func (rs *ResultSet) Configure(c *ConfigurationFile) {
	rs.config = c
	for i := range rs.results {
		rs.results[i] = rs.config.Apply(rs.results[i])
	}
}

// AddResult adds a result to the ResultSet, applying the current configuration if set
func (rs *ResultSet) AddResult(r ResultContext) {
	if rs.config != nil {
		r = rs.config.Apply(r)
	}
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
			return rule[i].Dashboard.Title < rule[j].Dashboard.Title
		})
	}
	return ret
}

func (rs *ResultSet) ReportByRule() {
	byRule := rs.ByRule()
	rules := make([]string, 0, len(byRule))
	for r := range byRule {
		rules = append(rules, r)
	}
	sort.Strings(rules)

	for _, rule := range rules {
		fmt.Fprintln(os.Stdout, byRule[rule][0].Rule.Description())
		for _, r := range byRule[rule] {
			if r.Result.Severity == Exclude && !rs.config.Verbose {
				continue
			}
			r.TtyPrint()
		}
	}
}
