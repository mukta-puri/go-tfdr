package config

import (
	"github.com/spf13/cobra"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
)

var newConfigCmd = &cobra.Command{
	Use:   "new",
	Short: "Generates a terraform state copy config file in $HOME/.tfstatecopy",
	Long:  `Generates a terraform state copy config config file in $HOME/.tfstatecopy`,
	Run: func(cmd *cobra.Command, args []string) {
		config.GenerateConfig()
	},
}

func init() {
	ConfigCmd.AddCommand(newConfigCmd)
}
