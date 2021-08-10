package api

import (
	"fmt"

	"github.com/mupuri/go-tfdr/internal/filter"
	"github.com/mupuri/go-tfdr/internal/tfdrerrors"
)

// DeleteTFStateResources &
func DeleteTFStateResources(workspaceName string, filterConfigFileName string) error {
	state, err := pullTFState(workspaceName)
	if err != nil {
		return tfdrerrors.ErrReadState{Err: err}
	}
	if state == nil {
		return tfdrerrors.ErrSourceIsEmpty{}
	}

	state.Resources, err = filter.StateFilter(state.Resources, filter.DeleteResourceFilterFunc, filterConfigFileName)
	if err != nil {
		return tfdrerrors.ErrUnableToFilter{Err: err}
	}
	state.Serial++

	err = createTFStateVersion(state, workspaceName)
	if err != nil {
		return fmt.Errorf("Unable to create new state version. Error: %v", err)
	}
	return nil

}
