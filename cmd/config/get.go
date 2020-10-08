package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
	"gopkg.in/yaml.v2"
)

var getConfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Display currently configured options",
	Long:  `Display currently configured options`,
	Run: func(cmd *cobra.Command, args []string) {
		bytes, _ := yaml.Marshal(config.GetConfig())
		fmt.Print(string(bytes))
	},
}

func init() {
	ConfigCmd.AddCommand(getConfigCmd)
}
