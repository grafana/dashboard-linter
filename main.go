package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations/lint"
)

var lintStrictFlag, lintByIntegrationFlag, lintVerboseFlag bool

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint all integrations",
	Long:  `Returns warnings or errors for integrations which do not adhere to accepted standards`,
	PreRun: func(cmd *cobra.Command, args []string) {
		_ = viper.BindPFlags(cmd.PersistentFlags())
	},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var warns []error
		minSeverity := lint.Error
		slug := viper.GetString(listSpecificSlugFlag)
		path := viper.GetString(integrationsPathFlag)
		rs := lint.NewRuleSet()
		is := loadIntegrations(path)
		config := lint.NewConfiguration()
		config.SetVerbose(lintVerboseFlag)
		if slug != "" {
			if i, ok := is[slug]; ok {
				rs.AddIntegration(i)
				cf, err := lint.LoadIntegrationLintConfig(i)
				if err != nil {
					warns = append(warns, err)
				} else {
					fmt.Println(cf)
					config.AddConfiguration(slug, cf)
				}
			}
		} else {
			for _, i := range is {
				rs.AddIntegration(i)
				cf, err := lint.LoadIntegrationLintConfig(i)
				if err != nil {
					warns = append(warns, err)
				} else {
					config.AddConfiguration(i.Meta.Slug, cf)
				}
			}
		}

		for _, w := range warns {
			fmt.Printf("WARN - Failed to load lint configuration: %s\n", w)
		}

		if lintStrictFlag {
			minSeverity = lint.Warning
		}

		res, err := rs.Lint()
		if err != nil {
			log.Fatalln("Problems during lint execution:", err)
		}

		res.Configure(config)
		if lintByIntegrationFlag {
			res.ReportByIntegration()
		} else {
			res.ReportByRule()
		}

		if res.MaximumSeverity() >= minSeverity {
			return fmt.Errorf("there were linting errors, please see previous output")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
	lintCmd.PersistentFlags().StringP(
		listSpecificSlugFlag,
		"s",
		"",
		"lint a specific integration slug",
	)
	lintCmd.PersistentFlags().StringP(
		integrationsPathFlag,
		"p",
		"",
		"integrations folder path",
	)
	lintCmd.Flags().BoolVar(
		&lintStrictFlag,
		"strict",
		false,
		"fail upon linting error or warning",
	)
	lintCmd.Flags().BoolVar(
		&lintByIntegrationFlag,
		"by-integration",
		false,
		"print results grouped by integration, rather than by rule",
	)
	lintCmd.Flags().BoolVar(
		&lintVerboseFlag,
		"verbose",
		false,
		"show more information about linting",
	)
}

var rootCmd = &cobra.Command{
	Use:   "integrations-ctl",
	Short: "A command-line application to interact with integrations",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

func loadIntegrations(path string) integrations.Integrations {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	is, err := integrations.LoadAll(logger, path)
	if err != nil {
		exitWithErr(fmt.Sprintf("Failed to load all integrations: %s", err))
	}
	return is
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exitWithErr(s string) {
	fmt.Fprintln(os.Stderr, s)
	os.Exit(1)
}

func prettyPrint(x interface{}) string {
	s, _ := json.MarshalIndent(x, "", "\t")
	return string(s)
}


func main() {
	cmd.Execute()
}

