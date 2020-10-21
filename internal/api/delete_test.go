package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/logging"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/testutils"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/tfdrerrors"
)

type DeleteSuite struct {
	suite.Suite
}

func (s *DeleteSuite) SetupSuite() {}

func (s *DeleteSuite) SetupTest() {
	os.Setenv("TF_TEAM_TOKEN", "test")
	os.Setenv("TF_ORG_NAME", "team")
	config.InitConfig("")
	logging.InitLogger()
}

func (s *DeleteSuite) TearDownTest() {
	os.Unsetenv("TF_TEAM_TOKEN")
	os.Unsetenv("TF_ORG_NAME")
}

func (s *DeleteSuite) TestDeleteTFState() {
	cases := []struct {
		wks               *testutils.TfeTestWks
		filterFile        string
		shouldErr         bool
		errValidationFunc func(error) bool
		errMessage        string
	}{
		{
			wks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutils.NewState(),
				CsvResponder: testutils.NewResponder("test", "state-versions", "https://state"),
				SvPostResponder: func(req *http.Request) (*http.Response, error) {
					state, err := testutils.DecodeStateFromBody(req)
					s.NoError(err)

					numFilters := 2

					s.Equal(testutils.DefaultTerraformVersion, state.TerraformVersion)
					s.Equal(testutils.DefaultVersion, state.Version)
					s.Equal(testutils.DefaultLineage, state.Lineage)
					s.Equal(testutils.DefaultSerial+1, state.Serial)
					s.Equal(testutils.DefaultNumResources()-numFilters-len(testutils.GlobalResources), len(state.Resources))

					resp, err := testutils.NewJSONResponse("test", "state-versions", "https://state")
					s.NoError(err)

					return resp, nil
				},
			},
			filterFile: "./testdata/filterConfig.json",
			shouldErr:  false,
			errMessage: "Test succesful delete state resources failed",
		},
		{
			wks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CsvResponder: httpmock.NewStringResponder(404, ""),
			},
			filterFile:        "",
			shouldErr:         true,
			errValidationFunc: func(err error) bool { return errors.Is(err, tfdrerrors.ErrSourceIsEmpty{}) },
			errMessage:        "Test delete error when source current state not found failed",
		},
		{
			wks: &testutils.TfeTestWks{
				Name:   "testNoWorkspace",
				Exists: false,
			},
			filterFile: "",
			shouldErr:  true,
			errValidationFunc: func(err error) bool {
				return errors.Is(err, tfdrerrors.ErrReadState{
					Err: tfdrerrors.ErrGetWorkspace{Err: tfe.ErrResourceNotFound},
				})
			},
			errMessage: "Test delete error when source workspace not found failed",
		},
		{
			wks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutils.NewState(),
				CsvResponder: testutils.NewResponder("test", "state-versions", "https://state"),
			},
			filterFile: "./testdata/not-found-filter.json",
			shouldErr:  true,
			errValidationFunc: func(err error) bool {
				return strings.Contains(err.Error(), "Unable to filter resources from state. Error:")
			},
			errMessage: "Test delete error when invalid filter file failed",
		},
		{
			wks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutils.NewState(),
				CsvResponder: testutils.NewResponder("test", "state-versions", "https://state"),
				// SvPostResponder: httpmock.NewErrorResponder(errors.New("Error creating state")),
			},
			filterFile: "./testdata/filterConfig.json",
			shouldErr:  true,
			errValidationFunc: func(err error) bool {
				return strings.Contains(err.Error(), "Unable to create new state version. Error:")
			},
			errMessage: "Test delete error when state create error failed",
		},
	}

	for _, c := range cases {
		httpmock.ActivateNonDefault(httpClient)
		httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/ping", httpmock.NewStringResponder(204, ""))
		err := testutils.SetupWksMockHTTPResponses(c.wks)
		s.NoError(err, c.errMessage)

		err = DeleteTFStateResources(c.wks.Name, c.filterFile)

		if c.shouldErr {
			s.Error(err, c.errMessage)
			if c.errValidationFunc != nil && err != nil {
				s.True(c.errValidationFunc(err), fmt.Sprintf("%v. Invalid error returned: %v", c.errMessage, err))
			}
		} else {
			s.NoError(err, c.errMessage)
		}
		httpmock.DeactivateAndReset()
	}
}

func TestDeleteSuite(t *testing.T) {
	suite.Run(t, new(DeleteSuite))
}
