package tfdrerrors

import "fmt"

type ErrReadFilterFile struct {
	Err error
}

func (errReadFilterFile ErrReadFilterFile) Error() string {
	return fmt.Sprintf("Unable to get workspace. Err: %v", errReadFilterFile.Err)
}
