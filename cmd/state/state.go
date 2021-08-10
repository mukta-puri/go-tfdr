package state

import (
	"github.com/mupuri/go-tfdr/cmd/state/copy"
	"github.com/mupuri/go-tfdr/cmd/state/delete"
	"github.com/spf13/cobra"
)

var StateCmd = &cobra.Command{
	Use:   "state",
	Short: "Modifies tf workspace state",
	Long:  `Modifies tf workspace state`,
}

func init() {
	StateCmd.AddCommand(copy.CopyStateCmd)
	StateCmd.AddCommand(delete.DeleteStateCmd)
}
