package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/suite"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/logging"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/testutil"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/tfdrerrors"
)

type CopyTestSuite struct {
	suite.Suite
}

func (s *CopyTestSuite) SetupSuite() {
	os.Setenv("TF_TEAM_TOKEN", "test")
	os.Setenv("TF_ORG_NAME", "team")
	config.InitConfig("")
	logging.InitLogger()
}

func (s *CopyTestSuite) SetupTest() {}

func (s *CopyTestSuite) TearDownTest() {}

func (s *CopyTestSuite) TestCopyTFState() {
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
			errMessage:        "Test copy error when source current state not found failed",
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
		httpmock.ActivateNonDefault(httpClient)
		httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/ping", httpmock.NewStringResponder(204, ""))
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
		httpmock.DeactivateAndReset()
	}
}
