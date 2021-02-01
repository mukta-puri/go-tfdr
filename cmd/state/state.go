package state

import (
	"github.com/spf13/cobra"
	"github.com/tyler-technologies/go-tfdr/cmd/state/copy"
	"github.com/tyler-technologies/go-tfdr/cmd/state/delete"
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
