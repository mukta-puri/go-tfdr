package config

import (
	"fmt"

	"github.com/mupuri/go-tfdr/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var getConfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Display currently configured options",
	Long:  `Display currently configured options`,
	Run: func(cmd *cobra.Command, args []string) {
		bytes, _ := yaml.Marshal(config.GetConfig())
		fmt.Println(string(bytes))
	},
}

func init() {
	ConfigCmd.AddCommand(getConfigCmd)
}
