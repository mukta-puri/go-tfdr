package api

import (
	"fmt"

	"github.com/mupuri/go-tfdr/internal/filter"
	"github.com/mupuri/go-tfdr/internal/models"
	"github.com/mupuri/go-tfdr/internal/tfdrerrors"
)

// CopyTFState &
func CopyTFState(origWorkspaceName string, newWorkspaceName string, filterConfigFileName string) error {
	oldState, err := pullTFState(origWorkspaceName)
	if err != nil {
		return tfdrerrors.ErrReadState{Err: err}
	}
	if oldState == nil {
		return tfdrerrors.ErrSourceIsEmpty{}
	}

	newResources, err := filter.StateFilter(oldState.Resources, filter.CopyResourceFilterFunc, filterConfigFileName)
	if err != nil {
		return fmt.Errorf("Unable to filter resources from state. Error: %v", err)
	}

	newState, err := pullTFState(newWorkspaceName)
	if err != nil {
		return tfdrerrors.ErrReadState{Err: err}
	}
	if newState != nil {
		return tfdrerrors.ErrDestinationNotEmpty{}
	}

	newState = &models.State{
		TerraformVersion: oldState.TerraformVersion,
		Version:          oldState.Version,
		Resources:        newResources,
		Serial:           1,
	}

	err = createTFStateVersion(newState, newWorkspaceName)
	if err != nil {
		return tfdrerrors.ErrUnableToCreateStateVersion{Err: err}
	}

	return nil
}
