package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

// Configuration contains a map of ConfigurationFile, where the key is the integration slug
type Configuration struct {
	configs map[string]ConfigurationFile
}

// ConfigurationFile contains a map for rule exclusions, and warnings, where the key is the
// rule name to be excluded or downgraded to a warning
type ConfigurationFile struct {
	Exclusions map[string]*ConfigurationRuleEntries `yaml:"exclusions"`
	Warnings   map[string]*ConfigurationRuleEntries `yaml:"warnings"`
}

type ConfigurationRuleEntries struct {
	Reason  string               `json:"reason,omitempty"`
	Entries []ConfigurationEntry `json:"entries,omitempty"`
}

// ConfigurationEntry will exist precisely once for every instance of a rule violation you wish
// exclude or downgrade to a warning. Each ConfigurationEntry will have to be an *exact* match
// to the combination of attributes set. Reason will not be evaluated, and is an opportunity for
// the author to explain why the exception, or downgrade to warning exists.
type ConfigurationEntry struct {
	Reason    string `json:"reason,omitempty"`
	Dashboard string `json:"dashboard,omitempty"`
	Panel     string `json:"panel,omitempty"`
	// This gets (un)marshalled as a string, because a 0 index is valid, but also the zero value of an int
	TargetIdx string `json:"targetIdx"`
}

func (cre *ConfigurationRuleEntries) AddEntry(e ConfigurationEntry) {
	cre.Entries = append(cre.Entries, e)
}

func (ce *ConfigurationEntry) IsMatch(r ResultContext) bool {
	ret := true
	if r.Dashboard != nil && ce.Dashboard != r.Dashboard.Title {
		ret = false
	}

	if r.Panel != nil && ce.Panel != r.Panel.Title {
		ret = false
	}

	if r.Target != nil && ce.TargetIdx != "" {
		idx, err := strconv.Atoi(ce.TargetIdx)
		if err == nil && idx != r.Target.Idx {
			ret = false
		}
	}

	return ret
}

func (c *Configuration) AddConfiguration(slug string, cf ConfigurationFile) {
	c.configs[slug] = cf
}

func (c *Configuration) Apply(res ResultContext) ResultContext {
	if cByIntegration, ok := c.configs[res.Integration.Meta.Slug]; ok {
		if ces, ok := cByIntegration.Exclusions[res.Rule.Name()]; ok {
			matched := false
			if ces != nil {
				for _, ce := range ces.Entries {
					if ce.IsMatch(res) {
						matched = true
					}
				}
				if len(ces.Entries) == 0 {
					matched = true
				}
			} else {
				matched = true
			}
			if matched {
				res.Result.Severity = Exclude
				res.Result.Message = res.Result.Message + " (Excluded)"
			}
		}
		if ces, ok := cByIntegration.Warnings[res.Rule.Name()]; ok {
			matched := false
			if ces != nil {
				for _, ce := range ces.Entries {
					if ce.IsMatch(res) {
						matched = true
					}
				}
				if len(ces.Entries) == 0 {
					matched = true
				}
			} else {
				matched = true
			}
			if matched {
				res.Result.Severity = Warning
			}
		}
	}
	return res
}

func LoadIntegrationLintConfig(i *integrations.Integration) (ConfigurationFile, error) {
	var cf ConfigurationFile
	lintFilePath := filepath.Join(i.Meta.FilePath(), ".lint")
	f, err := os.Open(lintFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return cf, nil
		}
		return cf, err
	}

	dec := yaml.NewDecoder(f)
	err = dec.Decode(&cf)
	if err != nil {
		return cf, fmt.Errorf("could not unmarshal lint configuration %s: %w", lintFilePath, err)
	}
	return cf, nil
}

func NewConfiguration() *Configuration {
	return &Configuration{
		configs: make(map[string]ConfigurationFile),
	}
}
