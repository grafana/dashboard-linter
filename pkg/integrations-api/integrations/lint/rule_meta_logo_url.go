package lint

import (
	"fmt"
	"net/http"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

type MetaLogoURLRule struct {
	name string
}

func (r *MetaLogoURLRule) Description() string {
	return fmt.Sprintf("%s Checks the logo_url in the integration's metadata.", r.name)
}

func (r *MetaLogoURLRule) Lint(i *integrations.Integration) []Result {
	if i.Meta.LogoURL == "" {
		return []Result{
			{
				Severity:    Warning,
				Message:     "Metadata logo_url not set.",
				integration: i,
			},
		}
	}
	resp, err := http.Get(i.Meta.LogoURL)
	if err != nil {
		return []Result{
			{
				Severity:    Error,
				Message:     err.Error(),
				integration: i,
			},
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return []Result{
			{
				Severity:    Error,
				Message:     fmt.Sprintf("Failed to load %s with status code %d", i.Meta.LogoURL, resp.StatusCode),
				integration: i,
			},
		}
	}
	return []Result{
		{
			Severity:    Success,
			Message:     "OK",
			integration: i,
		},
	}
}

func NewMetaLogoURLRule() *MetaLogoURLRule {
	return &MetaLogoURLRule{
		name: "meta-logo-url",
	}
}
