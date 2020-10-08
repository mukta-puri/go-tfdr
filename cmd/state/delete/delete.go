package delete

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/api"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
)

var workspaceName string
var filterConfigFile string

// DeleteStateCmd &
var DeleteStateCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes selected resources from TF cloud workspace state",
	Long:  `Deletes selected resources from TF cloud workspace state`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(filterConfigFile) == 0 {
			return errors.New("filterConfigFile file is required")
		}
		if len(workspaceName) == 0 {
			return errors.New("workspaceName file is required")
		}
		return config.ValidateConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		api.DeleteTFState(workspaceName, filterConfigFile)
	},
}

func init() {
	DeleteStateCmd.PersistentFlags().StringVarP(&workspaceName, "workspaceName", "w", "", "workspace name")
	DeleteStateCmd.PersistentFlags().StringVarP(&filterConfigFile, "filterConfigFile", "f", "", "file with filter config with resources to copy")
}
