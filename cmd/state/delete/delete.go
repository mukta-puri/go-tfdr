package delete

import (
	"errors"

	"github.com/mupuri/go-tfdr/internal/api"
	"github.com/mupuri/go-tfdr/internal/config"
	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return api.DeleteTFStateResources(workspaceName, filterConfigFile)
	},
}

func init() {
	DeleteStateCmd.PersistentFlags().StringVarP(&workspaceName, "workspaceName", "w", "", "workspace name")
	DeleteStateCmd.PersistentFlags().StringVarP(&filterConfigFile, "filterConfigFile", "f", "", "file with filter config with resources to copy")
}
