package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	cfg "github.com/tyler-technologies/go-tfdr/cmd/config"
	state "github.com/tyler-technologies/go-tfdr/cmd/state"
	"github.com/tyler-technologies/go-tfdr/internal/config"
	"github.com/tyler-technologies/go-tfdr/internal/logging"
)

var rootCmd = &cobra.Command{
	Use:   "tfdr",
	Short: "Script for manipulating tf state during DR",
	Long:  `Script for manipulating tf workspace state during disaster recovery of environment`,
}

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Generate markdown documentation",
	Run: func(cmd *cobra.Command, args []string) {
		err := doc.GenMarkdownTree(rootCmd, "./docs")
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Execute will run the cli command
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.DisableAutoGenTag = true
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	rootCmd.AddCommand(cfg.ConfigCmd)
	rootCmd.AddCommand(state.StateCmd)
	rootCmd.AddCommand(docCmd)
}

func initConfig() {
	config.InitConfig(cfgFile)
	logging.InitLogger()
}
