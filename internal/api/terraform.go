package api

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-tfe"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/filter"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/tfdrerrors"
)

var httpClient = &http.Client{}

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

func createTFStateVersion(state *models.State, workspaceName string) error {
	c := config.GetConfig()

	tfeConfig := &tfe.Config{
		HTTPClient: httpClient,
		Token:      c.TerraformTeamToken,
	}

	client, err := tfe.NewClient(tfeConfig)
	if err != nil {
		return fmt.Errorf("Cannot create tfe client. Err: %v", err)
	}

	workspace, err := client.Workspaces.Read(context.Background(), c.TerraformOrgName, workspaceName)
	if err != nil {
		return tfdrerrors.ErrGetWorkspace{Err: err}
	}

	client.Workspaces.Lock(context.Background(), workspace.ID, tfe.WorkspaceLockOptions{})

	stateBytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal state object. Error: %v", err)
	}

	versionMd5Bytes := fmt.Sprintf("%x", md5.Sum(stateBytes))
	versionMd5 := string(versionMd5Bytes[:])
	serial := state.Serial

	base64State := base64.StdEncoding.EncodeToString(stateBytes)

	_, err = client.StateVersions.Create(context.Background(), workspace.ID, tfe.StateVersionCreateOptions{
		MD5:     &versionMd5,
		Serial:  &serial,
		State:   &base64State,
		Lineage: &state.Lineage,
	})
	if err != nil {
		return fmt.Errorf("Unable to create new state version. Err: %v", err)
	}
	client.Workspaces.Unlock(context.Background(), workspace.ID)
	return nil
}

func pullTFState(workspaceName string) (*models.State, error) {
	c := config.GetConfig()

	tfeConfig := &tfe.Config{
		HTTPClient: httpClient,
		Token:      c.TerraformTeamToken,
	}

	client, err := tfe.NewClient(tfeConfig)
	if err != nil {
		return nil, fmt.Errorf("Cannot create tfe client. Err: %v", err)
	}

	workspace, err := client.Workspaces.Read(context.Background(), c.TerraformOrgName, workspaceName)
	if err != nil {
		return nil, tfdrerrors.ErrGetWorkspace{Err: err}
	}

	sv, err := client.StateVersions.Current(context.Background(), workspace.ID)
	if err != nil {
		if err.Error() == tfe.ErrResourceNotFound.Error() {
			return nil, nil
		}
		return nil, tfdrerrors.ErrUnableToGetStateVersion{Err: err}
	}

	s, err := client.StateVersions.Download(context.Background(), sv.DownloadURL)
	if err != nil {
		return nil, tfdrerrors.ErrUnableToDownloadState{Err: err}
	}

	var state models.State

	err = json.Unmarshal(s, &state)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal downloaded state json. Err: : %v", err)
	}

	return &state, nil
}
