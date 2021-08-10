// +build test

package testutils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/mupuri/go-tfdr/internal/models"
)

var defaultNonGlobalResources int = 10

// DefaultTerraformVersion &
var (
	DefaultTerraformVersion string   = "0.13.4"
	DefaultLineage          string   = "test"
	DefaultVersion          int      = 4
	DefaultSerial           int64    = int64(1)
	GlobalResources         []string = []string{
		"aws_cloudfront_distribution",
		"aws_cloudfront_origin_access_identity",
		"aws_iam_access_key",
		"aws_iam_policy_document",
		"aws_iam_policy",
	}
)

// NewState &
func NewState() *models.State {
	return &models.State{
		Version:          DefaultVersion,
		TerraformVersion: DefaultTerraformVersion,
		Serial:           DefaultSerial,
		Lineage:          DefaultLineage,
		Outputs:          nil,
		Resources:        NewStateResources(),
	}
}

// DefaultNumResources &
func DefaultNumResources() int {
	return defaultNonGlobalResources + len(GlobalResources)
}

// NewStateResources &
func NewStateResources() []models.Resource {
	resources := make([]models.Resource, 0)

	for i := 0; i < 10; i++ {
		res := models.Resource{
			Module: fmt.Sprintf("module.test_module_%v", i),
			Mode:   "managed",
			Type:   fmt.Sprintf("type_%v", i),
			Name:   fmt.Sprintf("orig_name_%v", i),
			Instances: []models.Instance{
				{
					Attributes: map[string]interface{}{
						"attr1": "old_value_1",
						"attr2": "old_value_2",
					},
				},
			},
		}
		resources = append(resources, res)
	}

	for i, v := range GlobalResources {
		res := models.Resource{
			Module: fmt.Sprintf("module.test_global_module_%v", i),
			Mode:   "managed",
			Type:   v,
			Name:   fmt.Sprintf("global_orig_name_%v", i),
		}
		resources = append(resources, res)
	}

	return resources
}

func SetupWksMockHTTPResponses(wks *TfeTestWks) error {
	if wks != nil {
		if wks.Exists {
			httpmock.RegisterResponder(
				"GET",
				fmt.Sprintf("https://app.terraform.io/api/v2/organizations/team/workspaces/%v", wks.Name),
				NewResponder(wks.Name, "workspaces", ""),
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

func DecodeStateFromBody(req *http.Request) (models.State, error) {
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

func NewResponder(id, typ, hostedStateDownloadURL string) httpmock.Responder {
	res := newStateVersion(id, typ, hostedStateDownloadURL)

	return httpmock.NewJsonResponderOrPanic(200, res)
}

func NewJSONResponse(id, typ, hostedStateDownloadURL string) (*http.Response, error) {
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
