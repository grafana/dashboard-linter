package lint

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// https://grafana.com/docs/grafana/latest/variables/variable-types/global-variables/
var globalVariables = map[string]interface{}{
	"__rate_interval": "8869990787ms",
	"__interval":      "4867856611ms",
	"__interval_ms":   "7781188786",
	"__range_ms":      "6737667980",
	"__range_s":       "9397795485",
	"__range":         "6069770749ms",
	"__dashboard":     "AwREbnft",
	"__from":          time.Date(2020, 7, 13, 20, 19, 9, 254000000, time.UTC),
	"__to":            time.Date(2020, 7, 13, 20, 19, 9, 254000000, time.UTC),
	"__name":          "name",
	"__org":           42,
	"__org.name":      "orgname",
	"__user.id":       42,
	"__user.login":    "user",
	"__user.email":    "user@test.com",
	"timeFilter":      "time > now() - 7d",
	"__timeFilter":    "time > now() - 7d",
}

func expandVariables(expr string) string {
	re := regexp.MustCompile(`\$[[:word:]]+`)
	parts := strings.Split(expr, "\"")
	for i := range parts {
		if i%2 == 1 {
			continue
		}
		parts[i] = re.ReplaceAllStringFunc(parts[i], func(s string) string {
			return fmt.Sprintf("%s", globalVariables[strings.TrimLeft(s, "$")])
		})
	}
	fmt.Println(">>>>", strings.Join(parts, "\""))
	return strings.Join(parts, "\"")
}
