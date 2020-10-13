package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/logging"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/testutil"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/tfdrerrors"
)

type TestSuite struct {
	suite.Suite
}

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

func (s *TestSuite) SetupSuite() {
	os.Setenv("TF_TEAM_TOKEN", "test")
	os.Setenv("TF_ORG_NAME", "team")
	config.InitConfig("")
	logging.InitLogger()
}

func (s *TestSuite) SetupTest() {
	httpmock.ActivateNonDefault(httpClient)
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/ping", httpmock.NewStringResponder(204, ""))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/test", newResponder("test", "workspaces", ""))
}

func (s *TestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestCopyTFState() {
	cases := []struct {
		origwks           *tfeTestWks
		newwks            *tfeTestWks
		filterFile        string
		shouldErr         bool
		errValidationFunc func(error) bool
		errMessage        string
	}{
		{
			origwks: &tfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutil.NewState(),
				CsvResponder: newResponder("test", "state-versions", "https://state"),
			},
			newwks: &tfeTestWks{
				Name:         "test2",
				Exists:       true,
				CurrentState: nil,
				CsvResponder: httpmock.NewStringResponder(404, ""),
				SvPostResponder: func(req *http.Request) (*http.Response, error) {
					state, err := decodeStateFromBody(req)
					s.NoError(err)

					numFilters := 2

					s.Equal(testutil.DefaultTerraformVersion, state.TerraformVersion)
					s.Equal(testutil.DefaultVersion, state.Version)
					s.Equal("", state.Lineage)
					s.Equal(int64(1), state.Serial)
					s.Equal(numFilters+len(config.GlobalResources), len(state.Resources))

					resp, err := newJSONResponse("test2", "state-versions", "https://state")
					s.NoError(err)

					return resp, nil
				},
			},
			filterFile: "./testdata/filterConfig.json",
			shouldErr:  false,
			errMessage: "Test succesful copy state failed",
		},
		{
			origwks: &tfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutil.NewState(),
				CsvResponder: newResponder("test", "state-versions", "https://state"),
			},
			newwks: &tfeTestWks{
				Name:         "test2",
				Exists:       true,
				CsvResponder: newResponder("test2", "state-versions", "https://state"),
			},
			filterFile:        "./testdata/filterConfig.json",
			shouldErr:         true,
			errValidationFunc: func(err error) bool { return errors.Is(err, tfdrerrors.ErrDestinationNotEmpty{}) },
			errMessage:        "Test copy error when non empty destination state failed",
		},
		{
			origwks: &tfeTestWks{
				Name:         "test1",
				Exists:       true,
				CsvResponder: httpmock.NewStringResponder(404, ""),
			},
			newwks: &tfeTestWks{
				Name: "test2",
			},
			filterFile:        "",
			shouldErr:         true,
			errValidationFunc: func(err error) bool { return errors.Is(err, tfdrerrors.ErrSourceIsEmpty{}) },
			errMessage:        "Test copy error when source state not found failed",
		},
		{
			origwks: &tfeTestWks{
				Name:   "testNoWorkspace",
				Exists: false,
			},
			newwks: &tfeTestWks{
				Name: "test2",
			},
			filterFile: "",
			shouldErr:  true,
			errValidationFunc: func(err error) bool {
				return errors.Is(err, tfdrerrors.ErrReadState{
					Err: tfdrerrors.ErrGetWorkspace{Err: tfe.ErrResourceNotFound},
				})
			},
			errMessage: "Test copy error when source workspace not found failed",
		},
		{
			origwks: &tfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutil.NewState(),
				CsvResponder: newResponder("test", "state-versions", "https://state"),
			},
			newwks: &tfeTestWks{
				Name:         "test2",
				Exists:       true,
				CsvResponder: httpmock.NewStringResponder(404, ""),
			},
			filterFile: "./testdata/not-found-filter.json",
			shouldErr:  true,
			errValidationFunc: func(err error) bool {
				return strings.Contains(err.Error(), "Unable to filter resources from state. Error:")
			},
			errMessage: "Test copy error when invalid filter file failed",
		},
		{
			origwks: &tfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutil.NewState(),
				CsvResponder: newResponder("test", "state-versions", "https://state"),
			},
			newwks: &tfeTestWks{
				Name:            "test2",
				Exists:          true,
				CsvResponder:    httpmock.NewStringResponder(404, ""),
				SvPostResponder: httpmock.NewErrorResponder(errors.New("Error creating state")),
			},
			filterFile: "./testdata/filterConfig.json",
			shouldErr:  true,
			errValidationFunc: func(err error) bool {
				return strings.Contains(err.Error(), "Unable to create new state version. Error:")
			},
			errMessage: "Test copy error when state create error failed",
		},
	}

	for _, c := range cases {
		err := setupWksMockHTTPResponses(c.origwks)
		s.NoError(err, c.errMessage)
		err = setupWksMockHTTPResponses(c.newwks)
		s.NoError(err, c.errMessage)

		err = CopyTFState(c.origwks.Name, c.newwks.Name, c.filterFile)

		if c.shouldErr {
			s.Error(err, c.errMessage)
			if c.errValidationFunc != nil {
				s.True(c.errValidationFunc(err), fmt.Sprintf("%v. Invalid error returned: %v", c.errMessage, err))
			}
		} else {
			s.NoError(err, c.errMessage)
		}
	}
}

func (s *TestSuite) TestDeleteTFStateResources() {
	origState, err := json.Marshal(testutil.NewState())
	s.NoError(err)

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, string(origState)))

	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test/state-versions", func(req *http.Request) (*http.Response, error) {
		state, err := decodeStateFromBody(req)
		s.NoError(err)

		numFilters := 2

		s.Equal(testutil.DefaultTerraformVersion, state.TerraformVersion)
		s.Equal(testutil.DefaultVersion, state.Version)
		s.Equal(testutil.DefaultLineage, state.Lineage)
		s.Equal(testutil.DefaultSerial+1, state.Serial)
		s.Equal(testutil.DefaultNumResources()-numFilters-len(config.GlobalResources), len(state.Resources))

		resp, err := newJSONResponse("test", "state-versions", "https://state")
		s.NoError(err)

		return resp, nil
	})

	err = DeleteTFStateResources("test", "./testdata/filterConfig.json")
	s.NoError(err)
}

func (s *TestSuite) TestDeleteTFStateSorceIsEmpty() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", httpmock.NewStringResponder(404, ""))

	err := DeleteTFStateResources("test", "./testdata/filterConfig.json")

	s.Error(err)
	s.True(errors.Is(err, tfdrerrors.ErrSourceIsEmpty{}))
}

func (s *TestSuite) TestDeleteTFStateSorceWorkspaceNotFound() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/testNoWorkspace", httpmock.NewStringResponder(404, ""))

	err := DeleteTFStateResources("testNoWorkspace", "./testdata/filterConfig.json")

	s.Error(err)
	s.True(errors.Is(err, tfdrerrors.ErrReadState{
		Err: tfdrerrors.ErrGetWorkspace{Err: tfe.ErrResourceNotFound},
	}))
}

func (s *TestSuite) TestDeleteTFStateInvalidFilterFile() {
	origState, err := json.Marshal(testutil.NewState())
	s.NoError(err)

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, string(origState)))

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/test2", newResponder("test2", "workspaces", ""))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test2/current-state-version", httpmock.NewStringResponder(404, ""))

	err = DeleteTFStateResources("test", "./testdata/not-exist.json")
	s.Error(err)
	s.True(strings.Contains(err.Error(), "Unable to filter resources from state. Error:"))
}

func (s *TestSuite) TestDeleteTFStateCreateError() {
	origState, err := json.Marshal(testutil.NewState())
	s.NoError(err)

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, string(origState)))

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/test2", newResponder("test2", "workspaces", ""))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test2/current-state-version", httpmock.NewStringResponder(404, ""))
	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test2/state-versions", httpmock.NewErrorResponder(errors.New("Error creating state")))

	err = DeleteTFStateResources("test", "./testdata/filterConfig.json")
	s.Error(err)
	s.True(strings.Contains(err.Error(), "Unable to create new state version. Error:"))
}

func (s *TestSuite) TestCreateTFStateVersion() {
	state := testutil.NewState()
	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test/state-versions", newResponder("test", "state-versions", "https://state"))

	err := createTFStateVersion(state, "test")
	s.NoError(err)
}

func (s *TestSuite) TestCreateTFStateVersionNoWorkspace() {
	state := testutil.NewState()
	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test/state-versions", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/not-found", httpmock.NewStringResponder(404, ""))

	err := createTFStateVersion(state, "not-found")
	s.Error(err)
	s.True(errors.Is(err, tfdrerrors.ErrGetWorkspace{
		Err: tfe.ErrResourceNotFound,
	}))
}

func (s *TestSuite) TestPullTFState() {
	currentState, err := json.Marshal(testutil.NewState())
	s.NoError(err)

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, string(currentState)))
	st, err := pullTFState("test")
	s.NoError(err)
	s.NotNil(st)
	s.Equal(testutil.DefaultNumResources(), len(st.Resources))
}

func (s *TestSuite) TestPullTFStateNoState() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", httpmock.NewStringResponder(404, ""))
	st, err := pullTFState("test")
	s.NoError(err)
	s.Nil(st)
}

func (s *TestSuite) TestPullTFStateErrGetCurrentState() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", httpmock.NewErrorResponder(errors.New("Error getting current state version")))

	st, err := pullTFState("test")
	s.Error(err)
	s.True(strings.Contains(err.Error(), "Cannot get current state. Error:"))
	s.Nil(st)
}

func (s *TestSuite) TestPullTFStateErrDownloadState() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewErrorResponder(errors.New("Error downloading state")))
	st, err := pullTFState("test")
	s.Error(err)
	s.True(strings.Contains(err.Error(), "Cannot download state. Error:"))
	s.Nil(st)
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

func readFile(t *testing.T, filename string) []byte {
	file, err := os.Open(filename)
	assert.NoError(t, err)
	defer func() {
		err = file.Close()
		assert.NoError(t, err)
	}()
	b, err := ioutil.ReadAll(file)
	assert.NoError(t, err)
	return b
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
