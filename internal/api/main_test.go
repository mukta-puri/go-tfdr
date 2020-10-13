package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
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

type tfeTestWks struct {
	Name            string
	Exists          bool
	CurrentState    *models.State
	CsvResponder    httpmock.Responder
	SvPostResponder httpmock.Responder
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(CopyTestSuite))
	suite.Run(t, new(DeleteTestSuite))
	suite.Run(t, new(UtilTestSuite))
}

func setupWksMockHTTPResponses(wks *tfeTestWks) error {
	if wks != nil {
		if wks.Exists {
			httpmock.RegisterResponder(
				"GET",
				fmt.Sprintf("https://app.terraform.io/api/v2/organizations/team/workspaces/%v", wks.Name),
				newResponder(wks.Name, "workspaces", ""),
			)
			if wks.CsvResponder != nil {
				httpmock.RegisterResponder(
					"GET",
					fmt.Sprintf("https://app.terraform.io/api/v2/workspaces/%v/current-state-version", wks.Name),
					wks.CsvResponder,
				)
			}
			if wks.CurrentState != nil {
				wksStateJSON, err := json.Marshal(&wks.CurrentState)
				if err != nil {
					return err
				}

				httpmock.RegisterResponder(
					"GET",
					"https://state",
					httpmock.NewStringResponder(200, string(wksStateJSON)),
				)
			}
			if wks.SvPostResponder != nil {
				httpmock.RegisterResponder(
					"POST",
					fmt.Sprintf("https://app.terraform.io/api/v2/workspaces/%v/state-versions", wks.Name),
					wks.SvPostResponder,
				)
			}
		} else {
			httpmock.RegisterResponder(
				"GET",
				fmt.Sprintf("https://app.terraform.io/api/v2/organizations/team/workspaces/%v", wks.Name),
				httpmock.NewStringResponder(404, ""),
			)
		}
	}
	return nil
}

func decodeStateFromBody(req *http.Request) (models.State, error) {
	var sv stateVersion
	if err := json.NewDecoder(req.Body).Decode(&sv); err != nil {
		return models.State{}, fmt.Errorf("Invalid body. Err: %v", err)
	}

	stateBytes, err := base64.StdEncoding.DecodeString(sv.Data.Attributes.State)
	if err != nil {
		return models.State{}, fmt.Errorf("Invalid state. Cannot base64 decode. Err: %v", err)
	}

	var state models.State

	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		return models.State{}, fmt.Errorf("Invalid state. Cannot unmarshal. Err: %v", err)
	}

	return state, nil
}

func newResponder(id, typ, hostedStateDownloadURL string) httpmock.Responder {
	res := newStateVersion(id, typ, hostedStateDownloadURL)

	return httpmock.NewJsonResponderOrPanic(200, res)
}

func newJSONResponse(id, typ, hostedStateDownloadURL string) (*http.Response, error) {
	res := newStateVersion(id, typ, hostedStateDownloadURL)

	return httpmock.NewJsonResponse(200, res)
}

func newStateVersion(id, typ, hostedStateDownloadURL string) stateVersion {
	res := stateVersion{
		Data: stateVersionData{
			ID:  id,
			Typ: typ,
		},
	}
	if hostedStateDownloadURL != "" {
		res.Data.Attributes = responseAttr{
			HostedStateDownloadURL: hostedStateDownloadURL,
		}
	}

	return res
}
