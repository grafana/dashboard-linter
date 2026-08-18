package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/grafana/dashboard-linter/lint"
)

var lintStrictFlag bool
var lintVerboseFlag bool
var lintAutofixFlag bool
var lintReadFromStdIn bool
var lintConfigFlag string

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint [dashboard.json]",
	Short: "Lint a dashboard",
	Long:  `Returns warnings or errors for dashboard which do not adhere to accepted standards`,
	PreRun: func(cmd *cobra.Command, args []string) {
		_ = viper.BindPFlags(cmd.PersistentFlags())
	},
	SilenceUsage: true,
	Args:         cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var buf []byte
		var err error
		var filename string

		if lintReadFromStdIn {
			if lintAutofixFlag {
				return fmt.Errorf("can't read from stdin and autofix")
			}

			buf, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %v", err)
			}
		} else {
			if len(args) == 0 {
				return fmt.Errorf("no dashboard file specified, use 'lint <dashboard.json>' or 'lint --stdin'")
			}
			filename = args[0]
			buf, err = os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", filename, err)
			}
		}

		dashboard, err := lint.NewDashboard(buf)
		if err != nil {
			return fmt.Errorf("failed to parse dashboard: %v", err)
		}

		// if no config flag was passed, set a default path of a .lint file in the dashboards directory
		if lintConfigFlag == "" {
			lintConfigFlag = path.Join(path.Dir(filename), ".lint")
		}

		config := lint.NewConfigurationFile()
		if err := config.Load(lintConfigFlag); err != nil {
			return fmt.Errorf("failed to load lint config: %v", err)
		}
		config.Verbose = lintVerboseFlag
		config.Autofix = lintAutofixFlag

		rules := lint.NewRuleSet()
		results, err := rules.Lint([]lint.Dashboard{dashboard})
		if err != nil {
			return fmt.Errorf("failed to lint dashboard: %v", err)
		}

		if config.Autofix {
			changes := results.AutoFix(&dashboard)
			if changes > 0 {
				err = write(dashboard, filename, buf)
				if err != nil {
					return err
				}
			}
		}

		results.Configure(config)
		results.ReportByRule()

		if lintStrictFlag && results.MaximumSeverity() >= lint.Warning {
			return fmt.Errorf("there were linting errors, please see previous output")
		}
		return nil
	},
}

// write merges the linted dashboard back into the original file, preserving
// all fields the linter's model does not know about. Values from the linted
// dashboard win, arrays are replaced wholesale, and null values never
// overwrite existing content.
func write(dashboard lint.Dashboard, filename string, old []byte) error {
	newBytes, err := dashboard.Marshal()
	if err != nil {
		return err
	}

	var oldMap, newMap map[string]interface{}
	if err := json.Unmarshal(old, &oldMap); err != nil {
		return err
	}
	if err := json.Unmarshal(newBytes, &newMap); err != nil {
		return err
	}
	stripped := stripNulls(newMap)
	newMap, _ = stripped.(map[string]interface{})

	b, err := json.MarshalIndent(mergeMaps(oldMap, newMap), "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	out := strings.ReplaceAll(string(b), "\"options\": null,", "\"options\": [],")

	perm := os.FileMode(0600)
	if info, err := os.Stat(filename); err == nil {
		perm = info.Mode().Perm()
	}
	return os.WriteFile(filename, []byte(out), perm)
}

// stripNulls recursively removes null values and empty objects/arrays from a
// JSON tree, so that fields the linter's model does not populate (e.g. nil
// slices) do not end up polluting the fixed dashboard.
func stripNulls(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for k, val := range t {
			if val == nil {
				continue
			}
			if s := stripNulls(val); s != nil {
				out[k] = s
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case []interface{}:
		out := make([]interface{}, 0, len(t))
		for _, val := range t {
			if s := stripNulls(val); s != nil {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return v
	}
}

// mergeMaps recursively merges new into old: values from new win, arrays are
// replaced entirely, and a null value in new keeps the old value.
func mergeMaps(old, new map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(old))
	for k, v := range old {
		out[k] = v
	}
	for k, v := range new {
		if v == nil {
			continue
		}
		oldVal, ok := out[k]
		if !ok {
			out[k] = v
			continue
		}
		oldMap, oldIsMap := oldVal.(map[string]interface{})
		newMap, newIsMap := v.(map[string]interface{})
		if oldIsMap && newIsMap {
			out[k] = mergeMaps(oldMap, newMap)
			continue
		}
		out[k] = v
	}
	return out
}

var rulesCmd = &cobra.Command{
	Use:          "rules",
	Short:        "Print documentation about each lint rule.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		rules := lint.NewRuleSet()
		for _, rule := range rules.Rules() {
			_, _ = fmt.Fprintf(os.Stdout, "* `%s` - %s\n", rule.Name(), rule.Description())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(rulesCmd)
	lintCmd.Flags().BoolVar(
		&lintStrictFlag,
		"strict",
		false,
		"fail upon linting error or warning",
	)
	lintCmd.Flags().BoolVar(
		&lintVerboseFlag,
		"verbose",
		false,
		"show more information about linting",
	)
	lintCmd.Flags().BoolVar(
		&lintAutofixFlag,
		"fix",
		false,
		"automatically fix problems if possible",
	)
	lintCmd.Flags().StringVarP(
		&lintConfigFlag,
		"config",
		"c",
		"",
		"path to a configuration file",
	)
	lintCmd.Flags().BoolVar(
		&lintReadFromStdIn,
		"stdin",
		false,
		"read from stdin",
	)
}

var rootCmd = &cobra.Command{
	Use:   "dashboard-linter",
	Short: "A command-line application to lint Grafana dashboards.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
