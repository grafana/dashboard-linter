package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
		cleanRun := true
		minSeverity := lint.Error
		slug := viper.GetString(listSpecificSlugFlag)
		path := viper.GetString(integrationsPathFlag)
		is := loadIntegrations(path)
		rules := lint.NewRules()

		if viper.GetBool("strict") {
			minSeverity = lint.Warning
		}
		for _, r := range rules {
			fmt.Println(r.Description())
			for _, i := range is {
				if slug != "" && i.Meta.Slug != slug {
					continue
				}
				for _, res := range r.Lint(i) {
					res.TtyPrint()
					if res.Severity >= minSeverity {
						cleanRun = false
					}
				}
			}
		}
		if !cleanRun {
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
}
