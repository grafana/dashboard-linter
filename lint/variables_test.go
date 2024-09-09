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
			expr:   "rate(metric{}[$__rate_interval])",
			result: "rate(metric{}[211d12h44m22s50ms])",
		},
		{
			desc:   "Should support ${...} syntax",
			expr:   "rate(metric{}[${__rate_interval}])",
			result: "rate(metric{}[211d12h44m22s51ms])",
		},
		{
			desc:   "Should support [[...]] syntax",
			expr:   "rate(metric{}[[[__rate_interval]]])",
			result: "rate(metric{}[211d12h44m22s52ms])",
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
			result: "max by(bgludgvy_variable_csv_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:doublequote} syntax",
			expr:   "max by(${variable:doublequote}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_doublequote_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:glob} syntax",
			expr:   "max by(${variable:glob}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_glob_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:json} syntax",
			expr:   "max by(${variable:json}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_json_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:lucene} syntax",
			expr:   "max by(${variable:lucene}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_lucene_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:percentencode} syntax",
			expr:   "max by(${variable:percentencode}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_percentencode_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:pipe} syntax",
			expr:   "max by(${variable:pipe}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_pipe_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:raw} syntax",
			expr:   "max by(${variable:raw}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_raw_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:regex} syntax",
			expr:   "max by(${variable:regex}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_regex_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:singlequote} syntax",
			expr:   "max by(${variable:singlequote}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_singlequote_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:sqlstring} syntax",
			expr:   "max by(${variable:sqlstring}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_sqlstring_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:text} syntax",
			expr:   "max by(${variable:text}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_text_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc:   "Should support ${variable:queryparam} syntax",
			expr:   "max by(${variable:queryparam}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(bgludgvy_variable_queryparam_0) (rate(cpu{}[211d12h44m22s50ms]))",
		},
		{
			desc: "Should return an error for unknown syntax",
			expr: "max by(${a:b:c:d}) (rate(cpu{}[$__rate_interval]))",
			err:  fmt.Errorf("failed to parse expression: max by(${a:b:c:d}) (rate(cpu{}[211d12h44m22s50ms]))"),
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
			result: "max by(bgludgvy_var_0) (rate(cpu{}[211d12h44m22s68ms:211d12h44m22s71ms]))",
		},
		{
			desc: "Should recursively replace variables",
			expr: "sum (rate(cpu{}[$interval]))",
			variables: []Template{
				{Name: "interval", Current: map[string]interface{}{"value": "$__auto_interval_interval"}},
			},
			result: "sum (rate(cpu{}[211d12h44m22s68ms]))",
		},
		{
			desc: "Should support plain $__auto_interval, generated by grafonnet-lib (https://github.com/grafana/grafonnet-lib/blob/master/grafonnet/template.libsonnet#L100)",
			expr: "sum (rate(cpu{}[$interval]))",
			variables: []Template{
				{Name: "interval", Current: map[string]interface{}{"value": "$__auto_interval"}},
			},
			result: "sum (rate(cpu{}[211d12h44m22s68ms]))",
		},
	} {
		s, err := expandVariables(tc.expr, tc.variables)
		require.Equal(t, tc.err, err)
		require.Equal(t, tc.result, s, tc.desc)
	}
}

func TestReverseVariableExpansion(t *testing.T) {
	placeholderByValue = map[string]*placeholder{
		"bgludgvy_variable_queryparam_0":    {variable: "${variable:queryparam}", value: "bgludgvy_variable_queryparam_0"},
		"211d12h44m22s63ms":                 {variable: "${__range_s}", value: "211d12h44m22s63ms"},
		"211d12h44m22s66ms":                 {variable: "${__range}", value: "211d12h44m22s66ms"},
		"1294671549257.000":                 {variable: "$__to", value: "1294671549257.000"},
		"bgludgvy___name_0":                 {variable: "$__name", value: "bgludgvy___name_0"},
		"bgludgvy_variable_regex_0":         {variable: "${variable:regex}", value: "bgludgvy_variable_regex_0"},
		"bgludgvy_variable_glob_0":          {variable: "${variable:glob}", value: "bgludgvy_variable_glob_0"},
		"211d12h44m22s52ms":                 {variable: "[[__rate_interval]]", value: "211d12h44m22s52ms"},
		"211d12h44m22s56ms":                 {variable: "$__interval_ms", value: "211d12h44m22s56ms"},
		"bgludgvy___dashboard_2":            {variable: "[[__dashboard]]", value: "bgludgvy___dashboard_2"},
		"1294671549256.000":                 {variable: "[[__from]]", value: "1294671549256.000"},
		"bgludgvy___user.login_0":           {variable: "$__user.login", value: "bgludgvy___user.login_0"},
		"bgludgvy_var_1":                    {variable: "${var}", value: "bgludgvy_var_1"},
		"bgludgvy_var_2":                    {variable: "[[var]]", value: "bgludgvy_var_2"},
		"bgludgvy_variable_sqlstring_0":     {variable: "${variable:sqlstring}", value: "bgludgvy_variable_sqlstring_0"},
		"211d12h44m22s60ms":                 {variable: "${__range_ms}", value: "211d12h44m22s60ms"},
		"bgludgvy___org.name_2":             {variable: "[[__org.name]]", value: "bgludgvy___org.name_2"},
		"bgludgvy_variable_csv_0":           {variable: "${variable:csv}", value: "bgludgvy_variable_csv_0"},
		"bgludgvy_variable_lucene_0":        {variable: "${variable:lucene}", value: "bgludgvy_variable_lucene_0"},
		"bgludgvy_variable_percentencode_0": {variable: "${variable:percentencode}", value: "bgludgvy_variable_percentencode_0"},
		"bgludgvy___user.email_1":           {variable: "${__user.email}", value: "bgludgvy___user.email_1"},
		"211d12h44m22s53ms":                 {variable: "$__interval", value: "211d12h44m22s53ms"},
		"211d12h44m22s57ms":                 {variable: "${__interval_ms}", value: "211d12h44m22s57ms"},
		"211d12h44m22s58ms":                 {variable: "[[__interval_ms]]", value: "211d12h44m22s58ms"},
		"bgludgvy___dashboard_0":            {variable: "$__dashboard", value: "bgludgvy___dashboard_0"},
		"1294671549265.000":                 {variable: "[[__user.id]]", value: "1294671549265.000"},
		"bgludgvy___timeFilter_0":           {variable: "$__timeFilter", value: "bgludgvy___timeFilter_0"},
		"bgludgvy___timeFilter_1":           {variable: "${__timeFilter}", value: "bgludgvy___timeFilter_1"},
		"bgludgvy___timeFilter_2":           {variable: "[[__timeFilter]]", value: "bgludgvy___timeFilter_2"},
		"211d12h44m22s69ms":                 {variable: "${interval}", value: "211d12h44m22s69ms"},
		"1294671549268.000":                 {variable: "${__from:date:iso}", value: "1294671549268.000"},
		"211d12h44m22s54ms":                 {variable: "${__interval}", value: "211d12h44m22s54ms"},
		"211d12h44m22s64ms":                 {variable: "[[__range_s]]", value: "211d12h44m22s64ms"},
		"bgludgvy___name_2":                 {variable: "[[__name]]", value: "bgludgvy___name_2"},
		"bgludgvy___user.email_0":           {variable: "$__user.email", value: "bgludgvy___user.email_0"},
		"bgludgvy___user.email_2":           {variable: "[[__user.email]]", value: "bgludgvy___user.email_2"},
		"bgludgvy___dashboard_1":            {variable: "${__dashboard}", value: "bgludgvy___dashboard_1"},
		"1294671549261.000":                 {variable: "${__org}", value: "1294671549261.000"},
		"1294671549264.000":                 {variable: "${__user.id}", value: "1294671549264.000"},
		"bgludgvy_variable_pipe_0":          {variable: "${variable:pipe}", value: "bgludgvy_variable_pipe_0"},
		"bgludgvy_variable_raw_0":           {variable: "${variable:raw}", value: "bgludgvy_variable_raw_0"},
		"211d12h44m22s50ms":                 {variable: "$__rate_interval", value: "211d12h44m22s50ms"},
		"211d12h44m22s62ms":                 {variable: "$__range_s", value: "211d12h44m22s62ms"},
		"1294671549259.000":                 {variable: "[[__to]]", value: "1294671549259.000"},
		"bgludgvy___org.name_0":             {variable: "$__org.name", value: "bgludgvy___org.name_0"},
		"bgludgvy_timeFilter_1":             {variable: "${timeFilter}", value: "bgludgvy_timeFilter_1"},
		"bgludgvy___user.login_2":           {variable: "[[__user.login]]", value: "bgludgvy___user.login_2"},
		"1294671549267.000":                 {variable: "${__from:date}", value: "1294671549267.000"},
		"211d12h44m22s55ms":                 {variable: "[[__interval]]", value: "211d12h44m22s55ms"},
		"bgludgvy_var_0":                    {variable: "$var", value: "bgludgvy_var_0"},
		"1294671549269.000":                 {variable: "${__from:date:YYYY-MM}", value: "1294671549269.000"},
		"1294671549266.000":                 {variable: "${__from:date:seconds}", value: "1294671549266.000"},
		"bgludgvy_variable_singlequote_0":   {variable: "${variable:singlequote}", value: "bgludgvy_variable_singlequote_0"},
		"211d12h44m22s51ms":                 {variable: "${__rate_interval}", value: "211d12h44m22s51ms"},
		"211d12h44m22s61ms":                 {variable: "[[__range_ms]]", value: "211d12h44m22s61ms"},
		"211d12h44m22s67ms":                 {variable: "[[__range]]", value: "211d12h44m22s67ms"},
		"1294671549258.000":                 {variable: "${__to}", value: "1294671549258.000"},
		"bgludgvy___user.login_1":           {variable: "${__user.login}", value: "bgludgvy___user.login_1"},
		"211d12h44m22s70ms":                 {variable: "[[interval]]", value: "211d12h44m22s70ms"},
		"1294671549255.000":                 {variable: "${__from}", value: "1294671549255.000"},
		"bgludgvy_timeFilter_0":             {variable: "$timeFilter", value: "bgludgvy_timeFilter_0"},
		"bgludgvy_timeFilter_2":             {variable: "[[timeFilter]]", value: "bgludgvy_timeFilter_2"},
		"bgludgvy_variable_doublequote_0":   {variable: "${variable:doublequote}", value: "bgludgvy_variable_doublequote_0"},
		"bgludgvy_variable_text_0":          {variable: "${variable:text}", value: "bgludgvy_variable_text_0"},
		"bgludgvy___name_1":                 {variable: "${__name}", value: "bgludgvy___name_1"},
		"1294671549260.000":                 {variable: "$__org", value: "1294671549260.000"},
		"bgludgvy___org.name_1":             {variable: "${__org.name}", value: "bgludgvy___org.name_1"},
		"211d12h44m22s59ms":                 {variable: "$__range_ms", value: "211d12h44m22s59ms"},
		"1294671549254.000":                 {variable: "$__from", value: "1294671549254.000"},
		"1294671549262.000":                 {variable: "[[__org]]", value: "1294671549262.000"},
		"211d12h44m22s68ms":                 {variable: "$interval", value: "211d12h44m22s68ms"},
		"211d12h44m22s72ms":                 {variable: "${resolution}", value: "211d12h44m22s72ms"},
		"211d12h44m22s65ms":                 {variable: "$__range", value: "211d12h44m22s65ms"},
		"1294671549263.000":                 {variable: "$__user.id", value: "1294671549263.000"},
		"bgludgvy_variable_json_0":          {variable: "${variable:json}", value: "bgludgvy_variable_json_0"},
		"211d12h44m22s71ms":                 {variable: "$resolution", value: "211d12h44m22s71ms"},
		"211d12h44m22s73ms":                 {variable: "[[resolution]]", value: "211d12h44m22s73ms"},
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
			expr:   "rate(metric{}[211d12h44m22s50ms])",
			result: "rate(metric{}[$__rate_interval])",
		},
		{
			desc:   "Should support ${...} syntax",
			expr:   "rate(metric{}[211d12h44m22s51ms])",
			result: "rate(metric{}[${__rate_interval}])",
		},
		{
			desc:   "Should support [[...]] syntax",
			expr:   "rate(metric{}[211d12h44m22s52ms])",
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
			expr:   "max by(bgludgvy_variable_csv_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:csv}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:doublequote} syntax",
			expr:   "max by(bgludgvy_variable_doublequote_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:doublequote}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:glob} syntax",
			expr:   "max by(bgludgvy_variable_glob_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:glob}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:json} syntax",
			expr:   "max by(bgludgvy_variable_json_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:json}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:lucene} syntax",
			expr:   "max by(bgludgvy_variable_lucene_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:lucene}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:percentencode} syntax",
			expr:   "max by(bgludgvy_variable_percentencode_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:percentencode}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:pipe} syntax",
			expr:   "max by(bgludgvy_variable_pipe_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:pipe}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:raw} syntax",
			expr:   "max by(bgludgvy_variable_raw_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:raw}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:regex} syntax",
			expr:   "max by(bgludgvy_variable_regex_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:regex}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:singlequote} syntax",
			expr:   "max by(bgludgvy_variable_singlequote_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:singlequote}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:sqlstring} syntax",
			expr:   "max by(bgludgvy_variable_sqlstring_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:sqlstring}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:text} syntax",
			expr:   "max by(bgludgvy_variable_text_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:text}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should support ${variable:queryparam} syntax",
			expr:   "max by(bgludgvy_variable_queryparam_0) (rate(cpu{}[211d12h44m22s50ms]))",
			result: "max by(${variable:queryparam}) (rate(cpu{}[$__rate_interval]))",
		},
		{
			desc:   "Should replace variables present in the templating",
			expr:   "max by(bgludgvy_var_0) (rate(cpu{}[211d12h44m22s68ms:211d12h44m22s71ms]))",
			result: "max by($var) (rate(cpu{}[$interval:$resolution]))",
		},
		{
			desc:   "Should recursively replace variables",
			expr:   "sum (rate(cpu{}[211d12h44m22s68ms]))",
			result: "sum (rate(cpu{}[$interval]))",
		},
		{
			desc:   "Should support plain $__auto_interval, generated by grafonnet-lib (https://github.com/grafana/grafonnet-lib/blob/master/grafonnet/template.libsonnet#L100)",
			expr:   "sum (rate(cpu{}[211d12h44m22s68ms]))",
			result: "sum (rate(cpu{}[$interval]))",
		},
	} {
		s := revertExpandedVariables(tc.expr)
		require.Equal(t, tc.result, s, tc.desc)
	}
}
