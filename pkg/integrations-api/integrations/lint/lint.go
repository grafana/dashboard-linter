package lint

import (
	"fmt"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

type Severity int

const (
	Success Severity = iota
	Warning
	Error
)

type Result struct {
	Severity    Severity
	Message     string
	integration *integrations.Integration
}

func (r Result) TtyPrint() {
	var sym string
	switch s := r.Severity; s {
	case Success:
		sym = "✔️"
	case Warning:
		sym = "⚠️"
	case Error:
		sym = "❌"
	}

	fmt.Printf("[%s] Integration: %s - %s\n", sym, r.integration.Meta.Slug, r.Message)
}

type Rule interface {
	Lint(*integrations.Integration) []Result
	Description() string
}

func NewRules() []Rule {
	return []Rule{
		NewMetaLogoURLRule(),
		NewMetaSupportedPlatformsRule(),
	}
}
