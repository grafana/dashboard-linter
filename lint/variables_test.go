package lint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariableExpansion(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		expr   string
		result string
		err    error
	}{
		{
			desc:   "Should not replace variables in quoted strings",
			expr:   "up{job=~\"$job\"}",
			result: "up{job=~\"$job\"}",
		},
		// https://grafana.com/docs/grafana/latest/variables/syntax/
		{
			desc:   "Should replace variables in metric name",
			expr:   "up$var{job=~\"$job\"}",
			result: "upvar{job=~\"$job\"}",
		},
		{
			desc:   "Should replace global rate/range variables",
			expr:   "rate(metric{}[$__rate_interval])",
			result: "rate(metric{}[8869990787ms])",
		},
		{
			desc:   "Should support ${...} syntax",
			expr:   "rate(metric{}[${__rate_interval}])",
			result: "rate(metric{}[8869990787ms])",
		},
		{
			desc:   "Should support [[...]] syntax",
			expr:   "rate(metric{}[[[__rate_interval]]])",
			result: "rate(metric{}[8869990787ms])",
		},
		// https://grafana.com/docs/grafana/latest/variables/variable-types/global-variables/
		{
			desc:   "Should support ${__user.id}",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__user.id})",
			result: "sum(http_requests_total{method=\"GET\"} @ 42)",
		},
		{
			desc:   "Should support $__from/$__to",
			expr:   "sum(http_requests_total{method=\"GET\"} @ $__from)",
			result: "sum(http_requests_total{method=\"GET\"} @ 1594671549254)",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (unix seconds)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__from:date:seconds}000)",
			result: "sum(http_requests_total{method=\"GET\"} @ 1594671549000)",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (iso default)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__from:date})",
			result: "sum(http_requests_total{method=\"GET\"} @ 2020-07-13T20:19:09Z)",
		},
		{
			desc:   "Should support $__from/$__to with formatting option (iso)",
			expr:   "sum(http_requests_total{method=\"GET\"} @ ${__from:date:iso})",
			result: "sum(http_requests_total{method=\"GET\"} @ 2020-07-13T20:19:09Z)",
		},
		{
			desc: "Should not support $__from/$__to with momentjs formatting option (iso)",
			expr: "sum(http_requests_total{method=\"GET\"} @ ${__from:date:YYYY-MM})",
			err:  fmt.Errorf("Unsupported momentjs time format: YYYY-MM"),
		},
		// https://grafana.com/docs/grafana/latest/variables/advanced-variable-format-options/
		{
			desc:   "Should support ${variable:csv} syntax",
			expr:   "max by(${variable:csv}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(variable,variable,variable) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:doublequote} syntax",
			expr:   "max by(${variable:doublequote}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(\"variable\",\"variable\",\"variable\") (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:glob} syntax",
			expr:   "max by(${variable:glob}) (rate(cpu{}[$__rate_interval]))",
			result: "max by({variable,variable,variable}) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:json} syntax",
			expr:   "max by(${variable:json}) (rate(cpu{}[$__rate_interval]))",
			result: "max by([\"variable\",\"variable\",\"variable\"]) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:lucene} syntax",
			expr:   "max by(${variable:lucene}) (rate(cpu{}[$__rate_interval]))",
			result: "max by((\"variable\" OR \"variable\" OR \"variable\")) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:percentencode} syntax",
			expr:   "max by(${variable:percentencode}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(variable%2Cvariable%2Cvariable) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:pipe} syntax",
			expr:   "max by(${variable:pipe}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(variable|variable|variable) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:raw} syntax",
			expr:   "max by(${variable:raw}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(variable,variable,variable) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:regex} syntax",
			expr:   "max by(${variable:regex}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(variable|variable|variable) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:singlequote} syntax",
			expr:   "max by(${variable:singlequote}) (rate(cpu{}[$__rate_interval]))",
			result: "max by('variable','variable','variable') (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:sqlstring} syntax",
			expr:   "max by(${variable:sqlstring}) (rate(cpu{}[$__rate_interval]))",
			result: "max by('variable','variable','variable') (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:text} syntax",
			expr:   "max by(${variable:text}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(variable + variable + variable) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc:   "Should support ${variable:queryparam} syntax",
			expr:   "max by(${variable:queryparam}) (rate(cpu{}[$__rate_interval]))",
			result: "max by(var-variable=variable&var-variable=variable&var-variable=variable) (rate(cpu{}[8869990787ms]))",
		},
		{
			desc: "Should return an error for unknown syntax",
			expr: "max by(${a:b:c:d}) (rate(cpu{}[$__rate_interval]))",
			err:  fmt.Errorf("unknown variable format: a:b:c:d"),
		},
	} {
		s, err := expandVariables(tc.expr)
		require.Equal(t, tc.err, err)
		require.Equal(t, tc.result, s, tc.desc)
	}
}
