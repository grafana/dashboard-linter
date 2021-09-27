package lint

import (
	"fmt"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

type MetaSupportedPlatformsRule struct {
	name string
}

func (r *MetaSupportedPlatformsRule) Description() string {
	return fmt.Sprintf("%s Checks that the supported_platforms in the integration's metadata is set.", r.name)
}

func (r *MetaSupportedPlatformsRule) Name() string {
	return r.name
}

func (r *MetaSupportedPlatformsRule) LintIntegration(i *integrations.Integration) Result {
	// If the supported platforms contained an invalid value, it would fail to load
	// in the first place. Maybe make that check reusable so it can be done here as well?
	if len(i.Meta.SupportedPlatforms) == 0 {
		if i.Meta.Visible {
			return Result{
				Severity: Error,
				Message:  "Metadata supported_platforms not set.",
			}
		}
		return Result{
			Severity: Warning,
			Message:  "Metadata supported_platforms not set, but integration is not visible.",
		}
	}
	return Result{
		Severity: Success,
		Message:  "OK",
	}
}

func NewMetaSupportedPlatformsRule() *MetaSupportedPlatformsRule {
	return &MetaSupportedPlatformsRule{
		name: "meta-supported-platforms",
	}
}
