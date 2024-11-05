package lint

import (
	"fmt"

	"github.com/prometheus/prometheus/model/labels"
)

func checkForMatcher(selector []*labels.Matcher, name string, ty labels.MatchType, value string) error {
	var result error
	result = fmt.Errorf("%s selector not found", name)

	for _, matcher := range selector {
		if matcher.Name != name {
			continue
		}
		if matcher.Type == ty && matcher.Value == value {
			result = nil
			break
		}

		if matcher.Type != ty {
			result = fmt.Errorf("%s selector is %s, not %s", name, matcher.Type, ty)
		}

		if matcher.Value != value {
			result = fmt.Errorf("%s selector is %s, not %s", name, matcher.Value, value)
		}
	}

	return result
}
