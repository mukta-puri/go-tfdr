package api

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/filter"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
)

// CopyTFState &
func CopyTFState(origWorkspaceName string, newWorkspaceName string, filterConfigFileName string) error {
	oldState, err := pullTFState(origWorkspaceName)
	if err != nil {
		return fmt.Errorf("Unable to read origin state. Error: %v", err)
	}

	newResources := filter.StateFilter(oldState.Resources, filter.CopyResourceFilterFunc, filterConfigFileName)

	newState, err := pullTFState(newWorkspaceName)
	if newState.TerraformVersion != "" {
		return fmt.Errorf("new workspace state is not empty")
	}

	newState.TerraformVersion = oldState.TerraformVersion
	newState.Version = oldState.Version

	newState.Resources = newResources
	newState.Serial++

	err = createTFState(newState, newWorkspaceName)
	if err != nil {
		return fmt.Errorf("Unable to create new state version. Error: %v", err)
	}

	return nil
}

// DeleteTFState &
func DeleteTFState(workspaceName string, filterConfigFileName string) error {
	state, err := pullTFState(workspaceName)
	if err != nil {
		return fmt.Errorf("Unable to read origin state. Error: %v", err)
	}

	state.Resources = filter.StateFilter(state.Resources, filter.DeleteResourceFilterFunc, filterConfigFileName)
	state.Serial++

	err = createTFState(state, workspaceName)
	if err != nil {
		return fmt.Errorf("Unable to create new state version. Error: %v", err)
	}
	return nil

}

func createTFState(state models.State, workspaceName string) error {
	c := config.GetConfig()

	config := &tfe.Config{
		Token: c.TerraformTeamToken,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		return fmt.Errorf("Cannot create tfe client. Err: %v", err)
	}

	workspace, err := client.Workspaces.Read(context.Background(), c.TerraformOrgName, workspaceName)
	if err != nil {
		return fmt.Errorf("Unable to get workspace. Err: %v", err)
	}

	client.Workspaces.Lock(context.Background(), workspace.ID, tfe.WorkspaceLockOptions{})

	stateBytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal state object. Err: %v", err)
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

func pullTFState(workspaceName string) (models.State, error) {
	c := config.GetConfig()

	config := &tfe.Config{
		Token: c.TerraformTeamToken,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		return models.State{}, fmt.Errorf("Cannot create tfe client. Err: %v", err)
	}

	workspace, err := client.Workspaces.Read(context.Background(), c.TerraformOrgName, workspaceName)
	if err != nil {
		return models.State{}, fmt.Errorf("Unable to get workspace. Err: %v", err)
	}

	sv, err := client.StateVersions.Current(context.Background(), workspace.ID)
	if err != nil {
		if err.Error() == "resource not found" {
			return models.State{}, nil
		}
		return models.State{}, fmt.Errorf("Cannot get current state. Err: %v", err)
	}

	s, err := client.StateVersions.Download(context.Background(), sv.DownloadURL)
	if err != nil {
		return models.State{}, fmt.Errorf("Cannot download state. Err: %v", err)
	}

	var state models.State

	err = json.Unmarshal(s, &state)
	if err != nil {
		return models.State{}, fmt.Errorf("Cannot unmarshal downloaded state json. Err: : %v", err)
	}

	return state, nil
}
