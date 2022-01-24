package lint

import (
	"fmt"
	"time"
)

const (
	stateInit = iota
	stateQuotedString
	stateVariable
	stateVariableCurly
	stateVariableSquareFirst
	stateVariableSquare
	stateError
)

// https://grafana.com/docs/grafana/latest/variables/variable-types/global-variables/
var globalVariables = map[string]interface{}{
	"__rate_interval": 8869990787 * time.Millisecond,
	"__interval":      4867856611 * time.Millisecond,
	"__interval_ms":   7781188786,
	"__range_ms":      6737667980,
	"__range_s":       9397795485,
	"__range":         6069770749 * time.Millisecond,
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
	out := make([]rune, 0, len(expr))
	state := stateInit // init
	for _, c := range expr {
		if state == -1 {
			return ""
		}
		fmt.Println(string(c), state)
		switch c {
		case '$':
			switch state {
			case stateInit:
				state = stateVariable
			default:
				state = stateError
			}
		case '"':
			switch state {
			case stateInit:
				state = stateQuotedString
			case stateQuotedString:
				state = stateInit
			}
			out = append(out, c)
		case '{':
			switch state {
			case stateInit:
				state = stateVariableCurly
			}
		case '}':
			switch state {
			case stateVariableCurly:
				state = stateInit
			}
		default:
			if state == stateInit || state == stateQuotedString {
				out = append(out, c)
			}
		}
	}
	return string(out)
}
