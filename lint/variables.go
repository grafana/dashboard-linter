package lint

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
)

const (
	rateInterval = "__rate_interval"
	interval     = "__interval"
	intervalMs   = "__interval_ms"
	rangeMs      = "__range_ms"
	rangeS       = "__range_s"
	rangeVar     = "__range"
	dashboard    = "__dashboard"
	from         = "__from"
	to           = "__to"
	name         = "__name"
	org          = "__org"
	orgName      = "__org.name"
	userID       = "__user.id"
	userLogin    = "__user.login"
	userEmail    = "__user.email"
	timeFilter   = "timeFilter"
	timeFilter2  = "__timeFilter"
	// magicTimeRange = model.Duration(time.Hour*24*211 + time.Hour*12 + time.Minute*44 + time.Second*22 + time.Millisecond*50) // 211d12h44m22s50ms
	magicTimeRange = 11277964 // seconds 130d12h46m4s
	magicEpoch     = float64(1294671549254)
	magicString    = "bgludgvy"
)

const (
	valTypeString valType = iota
	valTypeTimeRange
	valTypeEpoch
)

type valType int

type placeholder struct {
	variable string // variable including the "variable syntax" i.e. $var, ${var}, [[var]]
	valType  valType
	value    string
}

var placeholderByVariable = make(map[string]*placeholder)
var placeholderByValue = make(map[string]*placeholder)

var globalVariablesInit = false

// list of global variables in the for om a list of placeholders
var globalVariables = []*placeholder{
	{
		variable: rateInterval,
		valType:  valTypeTimeRange,
	},
	{
		variable: interval,
		valType:  valTypeTimeRange,
	},
	{
		variable: intervalMs,
		valType:  valTypeTimeRange,
	},
	{
		variable: rangeMs,
		valType:  valTypeTimeRange,
	},
	{
		variable: rangeS,
		valType:  valTypeTimeRange,
	},
	{
		variable: rangeVar,
		valType:  valTypeTimeRange,
	},
	{
		variable: dashboard,
		valType:  valTypeString,
	},
	{
		variable: from,
		valType:  valTypeEpoch,
	},
	{
		variable: to,
		valType:  valTypeEpoch,
	},
	{
		variable: name,
		valType:  valTypeString,
	},
	{
		variable: org,
		valType:  valTypeEpoch, // not really an epoch, but it is a float64
	},
	{
		variable: orgName,
		valType:  valTypeString,
	},
	{
		variable: userID,
		valType:  valTypeEpoch, // not really an epoch, but it is a float64
	},
	{
		variable: userLogin,
		valType:  valTypeString,
	},
	{
		variable: userEmail,
		valType:  valTypeString,
	},
	{
		variable: timeFilter,
		valType:  valTypeString, // not really a string, but currently we do only support prometheus queries, and this would not be a valid prometheus query...
	},
	{
		variable: timeFilter2,
		valType:  valTypeString, // not really a string, but currently we do only support prometheus queries, and this would not be a valid prometheus query...
	},
}

// var supportedFormatOptions = []string{"csv", "distributed", "doublequote", "glob", "json", "lucene", "percentencode", "pipe", "raw", "regex", "singlequote", "sqlstring", "text", "queryparam"}

var variableRegexp = regexp.MustCompile(
	strings.Join([]string{
		`("\$|\$)([[:word:]]+)`, // $var syntax
		`("\$|\$)\{([^}]+)\}`,   // ${var} syntax
		`\[\[([^\[\]]+)\]\]`,    // [[var]] syntax
	}, "|"),
)

func expandVariables(expr string, variables []Template) (string, error) {
	// initialize global variables if not already initialized
	if !globalVariablesInit {
		for _, v := range globalVariables {
			// assign placeholder to global variable 3 times to account for the 3 different ways a variable can be defined
			// $var, ${var}, [[var]]
			p := []placeholder{
				{
					variable: fmt.Sprintf("$%s", v.variable),
					valType:  v.valType,
				},
				{
					variable: fmt.Sprintf("${%s}", v.variable),
					valType:  v.valType,
				},
				{
					variable: fmt.Sprintf("[[%s]]", v.variable),
					valType:  v.valType,
				},
			}
			for _, v := range p {
				createPlaceholder(v.variable, v.valType)
			}
		}
		globalVariablesInit = true
	}
	// add template variables to placeholder maps
	for _, v := range variables {
		if v.Name != "" {
			// create placeholder 3 times to account for the 3 different ways a variable can be defined
			// at this point, we do not care about the value of the variable, we just need a placeholder for it.
			valType := getValueType(getTemplateVariableValue(v))
			createPlaceholder(fmt.Sprintf("$%s", v.Name), valType)
			createPlaceholder(fmt.Sprintf("${%s}", v.Name), valType)
			createPlaceholder(fmt.Sprintf("[[%s]]", v.Name), valType)
		}
	}

	expr = variableRegexp.ReplaceAllStringFunc(expr, RegexpExpandVariables)

	// Check if the expression can be parsed
	_, err := parser.ParseExpr(expr)
	if err != nil {
		// not using promql parser error since it contains memory address which is hard to test...
		return "", fmt.Errorf("failed to parse expression: %s", expr)
	}

	return expr, nil
}

func revertExpandedVariables(expr string) (string, error) {
	for _, p := range placeholderByValue {
		if p.valType == valTypeTimeRange {
			// Replace all versions of time range placeholder
			expr = strings.ReplaceAll(expr, p.value, p.variable)

			// Parse time duration
			d, err := model.ParseDuration(p.value + "s")
			if err != nil {
				return "", fmt.Errorf("failed to parse duration: %s when reverting expanded variable: %s", p.value, p.variable)
			}
			expr = strings.ReplaceAll(expr, d.String(), p.variable)

			// Parse as float64
			f, err := strconv.ParseFloat(p.value, 64)
			if err != nil {
				return "", fmt.Errorf("failed to parse float64: %s when reverting expanded variable: %s", p.value, p.variable)
			}
			expr = strings.ReplaceAll(expr, fmt.Sprint(f), p.variable)
		} else {
			expr = strings.ReplaceAll(expr, p.value, p.variable)
		}
	}
	return expr, nil
}

// Should not replace variables inside double quotes
func RegexpExpandVariables(s string) string {
	// check if string starts with a double quote
	if s[0:1] == `"` {
		return s
	}

	if strings.Contains(s, ":") {
		// check if variable is __from or __to with advanced formatting
		if strings.HasPrefix(trimVariableSyntax(s), from) || strings.HasPrefix(trimVariableSyntax(s), to) {
			if strings.Count(s, ":") > 2 {
				// Should not replace variables with more than 2 colons returning the original string, promql parser will handle the error.
				return s
			}
			return createPlaceholder(s, valTypeEpoch)
		}
		// check if variable contains more than 1 colon
		if strings.Count(s, ":") > 1 {
			// Should not replace variables with more than 1 colon returning the original string, promql parser will handle the error.
			return s
		}
	}
	return createPlaceholder(s, valTypeString)
}

// getPlaceholder returns placeholder for a provided variable or value
func getPlaceholder(variable string, value string) *placeholder {
	switch {
	case variable != "" && value != "":
		if p, ok := placeholderByVariable[variable]; ok {
			if p.value == value {
				return p
			}
		}
	case variable != "":
		if p, ok := placeholderByVariable[variable]; ok {
			return p
		}
	case value != "":
		if p, ok := placeholderByValue[value]; ok {
			return p
		}
	}
	return nil
}

// assignPlaceholder assigns a placeholder to a variable it ensures both placeholderByVariable and placeholderByValue are updated
func assignPlaceholder(placeholder placeholder) error {
	if placeholder.variable == "" || placeholder.value == "" {
		return fmt.Errorf("variable and value must not be empty")
	}
	// Check if variable and value combination already exists
	if getPlaceholder(placeholder.variable, placeholder.value) != nil {
		return nil
	}
	// check if value already exists but with a different variable
	p := getPlaceholder("", placeholder.value)
	if p != nil {
		if p.variable != placeholder.variable {
			return fmt.Errorf("value %s already assigned to variable %s", placeholder.value, p.variable)
		}
	}
	// check if variable already exists but with a different value
	p = getPlaceholder(placeholder.variable, "")
	if p != nil {
		if p.value != placeholder.value {
			return fmt.Errorf("variable %s already assigned to value %s", placeholder.variable, p.value)
		}
	}
	// add placeholder to placeholderByVariable
	placeholderByVariable[placeholder.variable] = &placeholder
	// add placeholder to placeholderByValue
	placeholderByValue[placeholder.value] = &placeholder
	return nil
}

func getValueType(value string) valType {
	// value might be provided as an integer, so we need to check if it can be parsed as an integer and then add s to the end)
	if _, err := strconv.Atoi(value); err == nil {
		value = value + "s"
	}
	// check if variable is a time range
	if _, err := model.ParseDuration(value); err == nil {
		return valTypeTimeRange
	}
	// check if variable is epoch, this is used for promql @ modifier
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return valTypeEpoch
	}
	return valTypeString
}

// createPlaceholder returns a placeholder for a variable.
func createPlaceholder(variable string, valType valType) string {
	// check if variable already has a placeholder
	if p := getPlaceholder(variable, ""); p != nil {
		return p.value
	}
	// create placeholder
	counter := 0
	var value string
	for {
		if valType == valTypeTimeRange {
			// Using magicTimeRange as a seed for the placeholder
			timeRange := magicTimeRange + counter
			value = strconv.Itoa(timeRange)
		}
		if valType == valTypeEpoch {
			// Using magicEpoch as a seed for the placeholder
			epoch := magicEpoch + float64(counter)
			// trim epoch to 3 decimal places since that is the precision used in prometheus
			value = fmt.Sprintf("%.3f", epoch)
		}
		if valType == valTypeString {
			value = fmt.Sprintf("%s_%s_%d", magicString, trimVariableSyntax(variable), counter)
		}

		if _, ok := placeholderByValue[value]; !ok {
			err := assignPlaceholder(placeholder{variable: variable, valType: valType, value: value})
			if err == nil {
				return value
			}
		}
		counter++
		if counter > 10000 {
			// this should never happen... but just in case... lets panic...
			panic("createPlaceholder: counter > 10000 - this should never happen :(")
		}
	}
}

// Helper func to remove the variable syntax from a string
func trimVariableSyntax(s string) string {
	s = strings.TrimPrefix(s, "[[")
	s = strings.TrimPrefix(s, "${")
	s = strings.TrimPrefix(s, "$")

	s = strings.TrimSuffix(s, "]]")
	s = strings.TrimSuffix(s, "}")

	// replace all ":" with "_"
	s = strings.ReplaceAll(s, ":", "_")

	return s
}

// Helper func to check if string has variable syntax
func checkVariableSyntax(s string) bool {
	return strings.Contains(s, "$") || strings.Contains(s, "[[") || strings.Contains(s, "{")
}

// Helper func to get the value of a template variable
func getTemplateVariableValue(v Template) string {
	var value string
	// do not handle error
	c, _ := v.Current.Get()
	// check if variable has a value
	if c.Value == "" {
		if len(v.Options) > 0 {
			// Do not handle error
			o, _ := v.Options[0].Get()
			if o.Value != "" {
				value = o.Value
			}
		}
	} else {
		value = c.Value
	}
	// check value for variable syntax
	if checkVariableSyntax(value) {
		// lazy way of dealing with __auto_interval...
		if strings.HasPrefix(trimVariableSyntax(value), "__auto_interval") {
			// This will result in a placeholder with type timeRange
			value = "9001s"
		} else {
			// try to expand variable
			varValue := getPlaceholder(value, "")
			if varValue != nil {
				value = varValue.value
			}
		}
	}
	return value
}
