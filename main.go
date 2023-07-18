package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zeitlinger/conflate"

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

func write(dashboard lint.Dashboard, filename string, old []byte) error {
	newBytes, err := dashboard.Marshal()
	if err != nil {
		return err
	}
	c := conflate.New()
	err = c.AddData(old, newBytes)
	if err != nil {
		return err
	}
	b, err := c.MarshalJSON()
	if err != nil {
		return err
	}
	json := strings.ReplaceAll(string(b), "\"options\": null,", "\"options\": [],")

	return os.WriteFile(filename, []byte(json), 0600)
}

var rulesCmd = &cobra.Command{
	Use:          "rules",
	Short:        "Print documentation about each lint rule.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		rules := lint.NewRuleSet()
		for _, rule := range rules.Rules() {
			fmt.Fprintf(os.Stdout, "* `%s` - %s\n", rule.Name(), rule.Description())
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
