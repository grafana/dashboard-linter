package lint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariableExpansion(t *testing.T) {
	for _, tc := range []struct {
		desc      string
		expr      string
		variables []Template
		result    string
		err       error
	}{
		{
			desc:   "Should not replace variables in quoted strings",
			expr:   "up{job=~\"$job\"}",
			result: "up{job=~\"$job\"}",
		},
		// https: //grafana.com/docs/grafana/latest/variables/syntax/
		{
			desc:   "Should replace variables in metric name",
			expr:   "up$var{job=~\"$job\"}",
			result: "upbgludgvy_var_0{job=~\"$job\"}",
		},
		{
			desc:   "Should replace global rate/range variables",
			expr:   "rate(metric{}[11277964])",
			result: "rate(metric{}[11277964])",
		},
		{
			desc:   "Should support ${...} syntax",
			expr:   "rate(metric{}[${__rate_interval}])",
			result: "rate(metric{}[11277965])",
		},
		{
			desc:   "Should support [[...]] syntax",
			expr:   "rate(metric{}[[[__rate_interval]]])",
			result: "rate(metric{}[11277966])",
		},
		// https://grafana.com/docs/grafana/latest/variables/variable-types/global-variables/
		{
			desc:   "Should support ${__user.id}",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__user.id})",
			result: "sum(http_requests_total{method=\"GET\"} @ 1294671549264.000)",
		},
		{
			desc:   "Should support $__from/$__to",
			expr:   "sum(http_requests_total{method=\"GET\"} @ $__from)",
			result: "sum(http_requests_total{method=\"GET\"} @ 1294671549254.000)",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (unix seconds)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__from:date:seconds})",
			result: "sum(http_requests_total{method=\"GET\"} @ 1294671549266.000)",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (iso default)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__from:date})",
			result: "sum(http_requests_total{method=\"GET\"} @ 1294671549267.000)",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (iso)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__from:date:iso})",
			result: "sum(http_requests_total{method=\"GET\"} @ 1294671549268.000)",
		},
		{
			desc:   "Should not support $__from/$__to with momentjs formatting option (iso)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__from:date:YYYY-MM})",
			result: "sum(http_requests_total{method=\"GET\"} @ 1294671549269.000)",
		},
		// https://grafana.com/docs/grafana/latest/variables/advanced-variable-format-options/
		{
			desc:   "Should support ${variable:csv} syntax",
			expr:   "max by(${variable:csv}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_csv_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:doublequote} syntax",
			expr:   "max by(${variable:doublequote}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_doublequote_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:glob} syntax",
			expr:   "max by(${variable:glob}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_glob_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:json} syntax",
			expr:   "max by(${variable:json}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_json_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:lucene} syntax",
			expr:   "max by(${variable:lucene}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_lucene_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:percentencode} syntax",
			expr:   "max by(${variable:percentencode}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_percentencode_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:pipe} syntax",
			expr:   "max by(${variable:pipe}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_pipe_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:raw} syntax",
			expr:   "max by(${variable:raw}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_raw_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:regex} syntax",
			expr:   "max by(${variable:regex}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_regex_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:singlequote} syntax",
			expr:   "max by(${variable:singlequote}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_singlequote_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:sqlstring} syntax",
			expr:   "max by(${variable:sqlstring}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_sqlstring_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:text} syntax",
			expr:   "max by(${variable:text}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_text_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support ${variable:queryparam} syntax",
			expr:   "max by(${variable:queryparam}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_queryparam_0) (rate(cpu{}[11277964]))",
		},
		{
			desc:   "Should support using variables for multiplication",
			expr:   "sum(rate(foo[$__rate_interval])) * $__range_s",
			result: "sum(rate(foo[11277964])) * 11277976",
		},
		{
			desc: "Should return an error for unknown syntax",
			expr: "max by(${a:b:c:d}) (rate(cpu{}[$__rate_interval]))",
			err:  fmt.Errorf("failed to parse expression: max by(${a:b:c:d}) (rate(cpu{}[11277964]))"),
		},
		{
			desc: "Should replace variables present in the templating",
			expr: "max by($var) (rate(cpu{}[$interval:$resolution]))",
			variables: []Template{
				{
					Name: "interval",
					Options: []RawTemplateValue{
						map[string]interface{}{
							"value": "4h",
						},
					},
				},
				{
					Name: "resolution",
					Options: []RawTemplateValue{
						map[string]interface{}{
							"value": "5m",
						},
					}},
				{
					Name: "var",
					Type: "query",
					Current: map[string]interface{}{
						"value": "value",
					}},
			},
			result: "max by(bgludgvy_var_0) (rate(cpu{}[11277982:11277985]))",
		},
		{
			desc: "Should recursively replace variables",
			expr: "sum (rate(cpu{}[$interval]))",
			variables: []Template{
				{Name: "interval", Current: map[string]interface{}{"value": "$__auto_interval_interval"}},
			},
			result: "sum (rate(cpu{}[11277982]))",
		},
		{
			desc: "Should support plain $__auto_interval, generated by grafonnet-lib (https://github.com/grafana/grafonnet-lib/blob/master/grafonnet/template.libsonnet#L100)",
			expr: "sum (rate(cpu{}[$interval]))",
			variables: []Template{
				{Name: "interval", Current: map[string]interface{}{"value": "$__auto_interval"}},
			},
			result: "sum (rate(cpu{}[11277982]))",
		},
	} {
		s, err := expandVariables(tc.expr, tc.variables)
		require.Equal(t, tc.err, err)
		require.Equal(t, tc.result, s, tc.desc)
	}
}

func TestReverseVariableExpansion(t *testing.T) {
	placeholderByValue = map[string]*placeholder{
		"bgludgvy___org.name_1":             {variable: "${__org.name}", valType: 0, value: "bgludgvy___org.name_1"},
		"bgludgvy_variable_doublequote_0":   {variable: "${variable:doublequote}", valType: 0, value: "bgludgvy_variable_doublequote_0"},
		"11277965":                          {variable: "${__rate_interval}", valType: 1, value: "11277965"},
		"11277966":                          {variable: "[[__rate_interval]]", valType: 1, value: "11277966"},
		"1294671549258.000":                 {variable: "${__to}", valType: 2, value: "1294671549258.000"},
		"bgludgvy___name_1":                 {variable: "${__name}", valType: 0, value: "bgludgvy___name_1"},
		"bgludgvy_variable_csv_0":           {variable: "${variable:csv}", valType: 0, value: "bgludgvy_variable_csv_0"},
		"bgludgvy_variable_regex_0":         {variable: "${variable:regex}", valType: 0, value: "bgludgvy_variable_regex_0"},
		"11277981":                          {variable: "[[__range]]", valType: 1, value: "11277981"},
		"1294671549264.000":                 {variable: "${__user.id}", valType: 2, value: "1294671549264.000"},
		"bgludgvy_timeFilter_2":             {variable: "[[timeFilter]]", valType: 0, value: "bgludgvy_timeFilter_2"},
		"1294671549269.000":                 {variable: "${__from:date:YYYY-MM}", valType: 2, value: "1294671549269.000"},
		"1294671549266.000":                 {variable: "${__from:date:seconds}", valType: 2, value: "1294671549266.000"},
		"1294671549268.000":                 {variable: "${__from:date:iso}", valType: 2, value: "1294671549268.000"},
		"bgludgvy_variable_json_0":          {variable: "${variable:json}", valType: 0, value: "bgludgvy_variable_json_0"},
		"bgludgvy_variable_singlequote_0":   {variable: "${variable:singlequote}", valType: 0, value: "bgludgvy_variable_singlequote_0"},
		"1294671549256.000":                 {variable: "[[__from]]", valType: 2, value: "1294671549256.000"},
		"1294671549260.000":                 {variable: "$__org", valType: 2, value: "1294671549260.000"},
		"1294671549263.000":                 {variable: "$__user.id", valType: 2, value: "1294671549263.000"},
		"bgludgvy___timeFilter_0":           {variable: "$__timeFilter", valType: 0, value: "bgludgvy___timeFilter_0"},
		"bgludgvy_var_2":                    {variable: "[[var]]", valType: 0, value: "bgludgvy_var_2"},
		"11277977":                          {variable: "${__range_s}", valType: 1, value: "11277977"},
		"11277979":                          {variable: "$__range", valType: 1, value: "11277979"},
		"1294671549259.000":                 {variable: "[[__to]]", valType: 2, value: "1294671549259.000"},
		"bgludgvy___user.email_0":           {variable: "$__user.email", valType: 0, value: "bgludgvy___user.email_0"},
		"bgludgvy_variable_glob_0":          {variable: "${variable:glob}", valType: 0, value: "bgludgvy_variable_glob_0"},
		"bgludgvy_variable_lucene_0":        {variable: "${variable:lucene}", valType: 0, value: "bgludgvy_variable_lucene_0"},
		"11277983":                          {variable: "${interval}", valType: 1, value: "11277983"},
		"11277980":                          {variable: "${__range}", valType: 1, value: "11277980"},
		"1294671549262.000":                 {variable: "[[__org]]", valType: 2, value: "1294671549262.000"},
		"bgludgvy___timeFilter_1":           {variable: "${__timeFilter}", valType: 0, value: "bgludgvy___timeFilter_1"},
		"bgludgvy_var_0":                    {variable: "$var", valType: 0, value: "bgludgvy_var_0"},
		"bgludgvy_variable_pipe_0":          {variable: "${variable:pipe}", valType: 0, value: "bgludgvy_variable_pipe_0"},
		"11277964":                          {variable: "$__rate_interval", valType: 1, value: "11277964"},
		"11277969":                          {variable: "[[__interval]]", valType: 1, value: "11277969"},
		"bgludgvy___dashboard_1":            {variable: "${__dashboard}", valType: 0, value: "bgludgvy___dashboard_1"},
		"1294671549255.000":                 {variable: "${__from}", valType: 2, value: "1294671549255.000"},
		"bgludgvy_variable_raw_0":           {variable: "${variable:raw}", valType: 0, value: "bgludgvy_variable_raw_0"},
		"bgludgvy_variable_sqlstring_0":     {variable: "${variable:sqlstring}", valType: 0, value: "bgludgvy_variable_sqlstring_0"},
		"bgludgvy_variable_text_0":          {variable: "${variable:text}", valType: 0, value: "bgludgvy_variable_text_0"},
		"11277978":                          {variable: "[[__range_s]]", valType: 1, value: "11277978"},
		"bgludgvy___dashboard_2":            {variable: "[[__dashboard]]", valType: 0, value: "bgludgvy___dashboard_2"},
		"bgludgvy___user.email_2":           {variable: "[[__user.email]]", valType: 0, value: "bgludgvy___user.email_2"},
		"bgludgvy_timeFilter_1":             {variable: "${timeFilter}", valType: 0, value: "bgludgvy_timeFilter_1"},
		"11277972":                          {variable: "[[__interval_ms]]", valType: 1, value: "11277972"},
		"bgludgvy___name_0":                 {variable: "$__name", valType: 0, value: "bgludgvy___name_0"},
		"bgludgvy___name_2":                 {variable: "[[__name]]", valType: 0, value: "bgludgvy___name_2"},
		"bgludgvy_var_1":                    {variable: "${var}", valType: 0, value: "bgludgvy_var_1"},
		"11277971":                          {variable: "${__interval_ms}", valType: 1, value: "11277971"},
		"bgludgvy___user.login_0":           {variable: "$__user.login", valType: 0, value: "bgludgvy___user.login_0"},
		"bgludgvy___user.login_1":           {variable: "${__user.login}", valType: 0, value: "bgludgvy___user.login_1"},
		"bgludgvy_variable_queryparam_0":    {variable: "${variable:queryparam}", valType: 0, value: "bgludgvy_variable_queryparam_0"},
		"bgludgvy_timeFilter_0":             {variable: "$timeFilter", valType: 0, value: "bgludgvy_timeFilter_0"},
		"11277967":                          {variable: "$__interval", valType: 1, value: "11277967"},
		"11277974":                          {variable: "${__range_ms}", valType: 1, value: "11277974"},
		"bgludgvy___org.name_2":             {variable: "[[__org.name]]", valType: 0, value: "bgludgvy___org.name_2"},
		"bgludgvy___user.email_1":           {variable: "${__user.email}", valType: 0, value: "bgludgvy___user.email_1"},
		"11277985":                          {variable: "$resolution", valType: 1, value: "11277985"},
		"11277987":                          {variable: "[[resolution]]", valType: 1, value: "11277987"},
		"11277973":                          {variable: "$__range_ms", valType: 1, value: "11277973"},
		"11277976":                          {variable: "$__range_s", valType: 1, value: "11277976"},
		"bgludgvy___dashboard_0":            {variable: "$__dashboard", valType: 0, value: "bgludgvy___dashboard_0"},
		"bgludgvy___user.login_2":           {variable: "[[__user.login]]", valType: 0, value: "bgludgvy___user.login_2"},
		"bgludgvy___timeFilter_2":           {variable: "[[__timeFilter]]", valType: 0, value: "bgludgvy___timeFilter_2"},
		"11277984":                          {variable: "[[interval]]", valType: 1, value: "11277984"},
		"1294671549254.000":                 {variable: "$__from", valType: 2, value: "1294671549254.000"},
		"bgludgvy_variable_percentencode_0": {variable: "${variable:percentencode}", valType: 0, value: "bgludgvy_variable_percentencode_0"},
		"bgludgvy___org.name_0":             {variable: "$__org.name", valType: 0, value: "bgludgvy___org.name_0"},
		"11277986":                          {variable: "${resolution}", valType: 1, value: "11277986"},
		"11277970":                          {variable: "$__interval_ms", valType: 1, value: "11277970"},
		"11277975":                          {variable: "[[__range_ms]]", valType: 1, value: "11277975"},
		"1294671549257.000":                 {variable: "$__to", valType: 2, value: "1294671549257.000"},
		"1294671549265.000":                 {variable: "[[__user.id]]", valType: 2, value: "1294671549265.000"},
		"11277968":                          {variable: "${__interval}", valType: 1, value: "11277968"},
		"1294671549261.000":                 {variable: "${__org}", valType: 2, value: "1294671549261.000"},
		"1294671549267.000":                 {variable: "${__from:date}", valType: 2, value: "1294671549267.000"},
		"11277982":                          {variable: "$interval", valType: 1, value: "11277982"},
	}
	for _, tc := range []struct {
		desc   string
		expr   string
		result string
	}{
		{
			desc:   "Should not replace variables in quoted strings",
			expr:   "up{job=~\"$job\"}",
			result: "up{job=~\"$job\"}",
		},
		// https: //grafana.com/docs/grafana/latest/variables/syntax/
		{
			desc:   "Should replace variables in metric name",
			expr:   "upbgludgvy_var_0{job=~\"$job\"}",
			result: "up$var{job=~\"$job\"}",
		},
		{
			desc:   "Should replace global rate/range variables",
			expr:   "rate(metric{}[130d12h46m4s])",
			result: "rate(metric{}[$__rate_interval])",
		},
		{
			desc:   "Should support ${...} syntax",
			expr:   "rate(metric{}[130d12h46m5s])",
			result: "rate(metric{}[${__rate_interval}])",
		},
		{
			desc:   "Should support [[...]] syntax",
			expr:   "rate(metric{}[130d12h46m6s])",
			result: "rate(metric{}[[[__rate_interval]]])",
		},
		// https://grafana.com/docs/grafana/latest/variables/variable-types/global-variables/
		{
			desc:   "Should support ${__user.id}",
			expr:   "sum(http_requests_total{method=\"GET\"} @ 1294671549264.000)",
			result: "sum(http_requests_total{method=\"GET\"} @ ${__user.id})",
		},
		{
			desc:   "Should support $__from/$__to",
			expr:   "sum(http_requests_total{method=\"GET\"} @ 1294671549254.000)",
			result: "sum(http_requests_total{method=\"GET\"} @ $__from)",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (unix seconds)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ 1294671549266.000)",
			result: "sum(http_requests_total{method=\"GET\"} @ ${__from:date:seconds})",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (iso default)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ 1294671549267.000)",
			result: "sum(http_requests_total{method=\"GET\"} @ ${__from:date})",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (iso)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ 1294671549268.000)",
			result: "sum(http_requests_total{method=\"GET\"} @ ${__from:date:iso})",
		},
		{
			desc:   "Should not support $__from/$__to with momentjs formatting option (iso)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ 1294671549269.000)",
			result: "sum(http_requests_total{method=\"GET\"} @ ${__from:date:YYYY-MM})",
		},
		// https://grafana.com/docs/grafana/latest/variables/advanced-variable-format-options/
		{
			desc:   "Should support ${variable:csv} syntax",
			expr:   "max by(bgludgvy_variable_csv_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:csv}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:doublequote} syntax",
			expr:   "max by(bgludgvy_variable_doublequote_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:doublequote}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:glob} syntax",
			expr:   "max by(bgludgvy_variable_glob_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:glob}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:json} syntax",
			expr:   "max by(bgludgvy_variable_json_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:json}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:lucene} syntax",
			expr:   "max by(bgludgvy_variable_lucene_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:lucene}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:percentencode} syntax",
			expr:   "max by(bgludgvy_variable_percentencode_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:percentencode}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:pipe} syntax",
			expr:   "max by(bgludgvy_variable_pipe_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:pipe}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:raw} syntax",
			expr:   "max by(bgludgvy_variable_raw_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:raw}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:regex} syntax",
			expr:   "max by(bgludgvy_variable_regex_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:regex}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:singlequote} syntax",
			expr:   "max by(bgludgvy_variable_singlequote_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:singlequote}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:sqlstring} syntax",
			expr:   "max by(bgludgvy_variable_sqlstring_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:sqlstring}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:text} syntax",
			expr:   "max by(bgludgvy_variable_text_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:text}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:queryparam} syntax",
			expr:   "max by(bgludgvy_variable_queryparam_0) (rate(cpu{}[130d12h46m4s]))",
			result: "max by(${variable:queryparam}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should replace variables present in the templating",
			expr:   "max by(bgludgvy_var_0) (rate(cpu{}[130d12h46m22s:130d12h46m25s]))",
			result: "max by($var) (rate(cpu{}[$interval:$resolution]))",
		},
		{
			desc:   "Should support using variables for multiplication",
			expr:   "sum(rate(foo[130d12h46m4s])) * 11277976",
			result: "sum(rate(foo[$__rate_interval])) * $__range_s",
		},
		{
			desc:   "Should recursively replace variables",
			expr:   "sum (rate(cpu{}[130d12h46m22s]))",
			result: "sum (rate(cpu{}[$interval]))",
		},
		{
			desc:   "Should support plain $__auto_interval, generated by grafonnet-lib (https://github.com/grafana/grafonnet-lib/blob/master/grafonnet/template.libsonnet#L100)",
			expr:   "sum (rate(cpu{}[130d12h46m22s]))",
			result: "sum (rate(cpu{}[$interval]))",
		},
	} {
		s, _ := revertExpandedVariables(tc.expr)
		require.Equal(t, tc.result, s, tc.desc)
	}
}
