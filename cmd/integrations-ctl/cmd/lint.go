package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations/lint"
)

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
		if slug != "" {
			if i, ok := is[slug]; ok {
				warns = rs.AddIntegrations(map[string]*integrations.Integration{slug: i})
			}
		} else {
			warns = rs.AddIntegrations(is)
		}

		for _, w := range warns {
			fmt.Printf("WARN - Failed to load lint configuration: %s\n", w)
		}

		if viper.GetBool("strict") {
			minSeverity = lint.Warning
		}

		res := rs.Lint()
		if viper.GetBool("by-integration") {
			res.ReportByIntegration()
		} else {
			res.ReportByRule()
		}

		if res.MaximumSeverity() >= minSeverity {
			return fmt.Errorf("There were linting errors, please see previous output")
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
	lintCmd.PersistentFlags().Bool(
		"strict",
		false,
		"fail upon linting error or warning",
	)
	lintCmd.PersistentFlags().Bool(
		"by-integration",
		false,
		"print results grouped by integration, rather than by rule",
	)
}
