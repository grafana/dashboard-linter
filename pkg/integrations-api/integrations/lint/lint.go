package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

type Severity int

const (
	Success Severity = iota
	Exclude
	Warning
	Error
)

type Result struct {
	Severity    Severity
	Message     string
	Rule        Rule
	Integration *integrations.Integration
}

func (r Result) TtyPrint() {
	var sym string
	switch s := r.Severity; s {
	case Success:
		sym = "✔️"
	case Exclude:
		sym = "➖"
	case Warning:
		sym = "⚠️"
	case Error:
		sym = "❌"
	}

	fmt.Printf("[%s] Integration: %s - %s\n", sym, r.Integration.Meta.Slug, r.Message)
}

type Rule interface {
	Lint(*integrations.Integration) []Result
	Description() string
	Name() string
}

// Configuration of linting Exclusions and Warnings are string maps where the key is
// the name of the rule to exclude or be downgraded to a warning.
// In many cases there will not be a value, but some more complex lint rules may
// need further detail, such as the dashboard name, or file line number to exclude.
// These more granular exclusions will be documented in each rule that supports them.
type Configuration struct {
	Exclusions map[string]interface{} `yaml:"exclusions,omitempty"`
	Warnings   map[string]interface{} `yaml:"warnings,omitempty"`
}

type ResultSet struct {
	results []Result
}

func (rs *ResultSet) MaximumSeverity() Severity {
	retVal := Success
	for _, res := range rs.results {
		if res.Severity > retVal {
			retVal = res.Severity
		}
	}
	return retVal
}

func (rs *ResultSet) ByRule() map[string][]Result {
	ret := make(map[string][]Result)
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

func (rs *ResultSet) ByIntegration() map[string][]Result {
	ret := make(map[string][]Result)
	for _, res := range rs.results {
		ret[res.Integration.Meta.Slug] = append(ret[res.Integration.Meta.Slug], res)
	}
	return ret
}

func (rs *ResultSet) ReportByRule() {
	for _, res := range rs.ByRule() {
		fmt.Println(res[0].Rule.Description())
		for _, r := range res {
			r.TtyPrint()
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
			r.TtyPrint()
		}
	}
}

// RuleSet contains a list of linting rules, a list of integrations, and list of configurations
// which dictate the ultimate application of a rule to an integration based on the configs.
type RuleSet struct {
	configs      map[string]Configuration
	rules        []Rule
	integrations map[string]*integrations.Integration
}

func (s *RuleSet) Rules() []Rule {
	return s.rules
}

func (s *RuleSet) Integrations() map[string]*integrations.Integration {
	return s.integrations
}

func (s *RuleSet) Lint() ResultSet {
	var results []Result
	for _, r := range s.rules {
		for _, i := range s.integrations {
			res := s.lintWithConfig(r, i)
			results = append(results, res...)
		}
	}
	return ResultSet{results: results}
}

// lintWithConfig runs the specified rule against the specified integration, then applies appropriate decorations
// to the results as defined in the configuration.
func (s *RuleSet) lintWithConfig(r Rule, i *integrations.Integration) []Result {
	res := r.Lint(i)
	if lc, hasConfig := s.configs[i.Meta.Slug]; hasConfig {
		if _, exclude := lc.Exclusions[r.Name()]; exclude {
			for i := range res {
				res[i].Severity = Exclude
				res[i].Message = res[i].Message + " (Excluded)"
			}
		}
		if _, warn := lc.Warnings[r.Name()]; warn {
			for i := range res {
				res[i].Severity = Warning
			}
		}
	}
	return res
}

// AddIntegrations loads the lint configuration of each provided integration then adds it to the internal
// integrations map of the RuleSet, and returns a list of lint configuration loading errors.
// Duplicate integrations (unique by slug) will be overwritten.
func (s *RuleSet) AddIntegrations(i map[string]*integrations.Integration) []error {
	e := make([]error, 0)
	for k, v := range i {
		err := s.loadLintConfig(v)
		if err != nil {
			e = append(e, err)
		}
		s.integrations[k] = v
	}
	return e
}

func (s *RuleSet) loadLintConfig(i *integrations.Integration) error {
	lintFilePath := filepath.Join(i.Meta.FilePath(), ".lint")
	f, err := os.Open(lintFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var lc Configuration
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&lc)
	if err != nil {
		return fmt.Errorf("could not unmarshal lint configuration %s: %w", lintFilePath, err)
	}
	s.configs[i.Meta.Slug] = lc
	return nil
}

func NewRuleSet() RuleSet {
	return RuleSet{
		configs:      make(map[string]Configuration),
		rules:        NewRules(),
		integrations: make(map[string]*integrations.Integration),
	}
}

func NewRules() []Rule {
	return []Rule{
		NewMetaLogoURLRule(),
		NewMetaSupportedPlatformsRule(),
	}
}
