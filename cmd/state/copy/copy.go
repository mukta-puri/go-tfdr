package copy

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/tyler-technologies/go-tfdr/internal/api"
	"github.com/tyler-technologies/go-tfdr/internal/config"
)

var originalWorkspaceName string
var newWorkspaceName string
var filterConfigFile string

var CopyStateCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copies state from one workspace to another",
	Long:  `Copies state from one workspace to another`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(filterConfigFile) == 0 {
			return errors.New("filterConfigFile file is required")
		}
		if len(originalWorkspaceName) == 0 {
			return errors.New("originalWorkspaceName is required")
		}
		if len(newWorkspaceName) == 0 {
			return errors.New("newWorkspaceName is required")
		}
		return config.ValidateConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return api.CopyTFState(originalWorkspaceName, newWorkspaceName, filterConfigFile)
	},
}

func init() {
	CopyStateCmd.PersistentFlags().StringVarP(&originalWorkspaceName, "originalWorkspaceName", "o", "", "workspace to copy state from")
	CopyStateCmd.PersistentFlags().StringVarP(&newWorkspaceName, "newWorkspaceName", "n", "", "workspace to copy state to")
	CopyStateCmd.PersistentFlags().StringVarP(&filterConfigFile, "filterConfigFile", "f", "", "file with filter config with resources to copy")
}
