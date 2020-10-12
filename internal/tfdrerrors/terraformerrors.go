package tfdrerrors

import "fmt"

type ErrDestinationNotEmpty struct{}

func (ErrDestinationNotEmpty) Error() string {
	return "new workspace state is not empty"
}

type ErrSourceIsEmpty struct{}

func (ErrSourceIsEmpty) Error() string {
	return "existing workspace state is empty"
}

type ErrReadState struct {
	Err error
}

func (errReadState ErrReadState) Error() string {
	return fmt.Sprintf("Unable to read origin state. Error: %v", errReadState.Err)
}

type ErrGetWorkspace struct {
	Err error
}

func (errGetWorkspace ErrGetWorkspace) Error() string {
	return fmt.Sprintf("Unable to get workspace. Error: %v", errGetWorkspace.Err)
}

type ErrUnableToFilter struct {
	Err error
}

func (errUnableToFilter ErrUnableToFilter) Error() string {
	return fmt.Sprintf("Unable to filter resources from state. Error: %v", errUnableToFilter.Err)
}

type ErrUnableToCreateStateVersion struct {
	Err error
}

func (errUnableToCreateStateVersion ErrUnableToCreateStateVersion) Error() string {
	return fmt.Sprintf("Unable to create new state version. Error: %v", errUnableToCreateStateVersion.Err)
}

type ErrUnableToGetStateVersion struct {
	Err error
}

func (errUnableToGetStateVersion ErrUnableToGetStateVersion) Error() string {
	return fmt.Sprintf("Cannot get current state. Error: %v", errUnableToGetStateVersion.Err)
}

type ErrUnableToDownloadState struct {
	Err error
}

func (errUnableToDownloadState ErrUnableToDownloadState) Error() string {
	return fmt.Sprintf("Cannot download state. Error: %v", errUnableToDownloadState.Err)
}
