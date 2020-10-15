// +build test

package testutils

import (
	"github.com/jarcoal/httpmock"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
)

type stateVersion struct {
	Data stateVersionData `json:"data"`
}

type stateVersionData struct {
	ID         string       `json:"id"`
	Typ        string       `json:"type"`
	Attributes responseAttr `json:"attributes,omitempty"`
}

type responseAttr struct {
	HostedStateDownloadURL string `json:"hosted-state-download-url,omitempty"`
	State                  string `json:"state,omitempty"`
}

type TfeTestWks struct {
	Name            string
	Exists          bool
	CurrentState    *models.State
	CsvResponder    httpmock.Responder
	SvPostResponder httpmock.Responder
}
