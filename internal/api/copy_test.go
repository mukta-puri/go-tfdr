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
	"github.com/tyler-technologies/go-tfdr/internal/config"
	"github.com/tyler-technologies/go-tfdr/internal/logging"
	"github.com/tyler-technologies/go-tfdr/internal/testutils"
	"github.com/tyler-technologies/go-tfdr/internal/tfdrerrors"
)

type CopySuite struct {
	suite.Suite
}

func (s *CopySuite) SetupSuite() {}

func (s *CopySuite) SetupTest() {
	os.Setenv("TF_TEAM_TOKEN", "test")
	os.Setenv("TF_ORG_NAME", "team")
	config.InitConfig("")
	logging.InitLogger()
}

func (s *CopySuite) TearDownTest() {
	os.Unsetenv("TF_TEAM_TOKEN")
	os.Unsetenv("TF_ORG_NAME")
}

func (s *CopySuite) TestCopyTFState() {
	cases := []struct {
		origwks           *testutils.TfeTestWks
		newwks            *testutils.TfeTestWks
		filterFile        string
		shouldErr         bool
		errValidationFunc func(error) bool
		errMessage        string
	}{
		{
			origwks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutils.NewState(),
				CsvResponder: testutils.NewResponder("test", "state-versions", "https://state"),
			},
			newwks: &testutils.TfeTestWks{
				Name:         "test2",
				Exists:       true,
				CurrentState: nil,
				CsvResponder: httpmock.NewStringResponder(404, ""),
				SvPostResponder: func(req *http.Request) (*http.Response, error) {
					state, err := testutils.DecodeStateFromBody(req)
					s.NoError(err)

					numFilters := 2

					s.Equal(testutils.DefaultTerraformVersion, state.TerraformVersion)
					s.Equal(testutils.DefaultVersion, state.Version)
					s.Equal("", state.Lineage)
					s.Equal(int64(1), state.Serial)
					s.Equal(numFilters+len(testutils.GlobalResources), len(state.Resources))

					resp, err := testutils.NewJSONResponse("test2", "state-versions", "https://state")
					s.NoError(err)

					return resp, nil
				},
			},
			filterFile: "./testdata/filterConfig.json",
			shouldErr:  false,
			errMessage: "Test succesful copy state failed",
		},
		{
			origwks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutils.NewState(),
				CsvResponder: testutils.NewResponder("test", "state-versions", "https://state"),
			},
			newwks: &testutils.TfeTestWks{
				Name:         "test2",
				Exists:       true,
				CsvResponder: testutils.NewResponder("test2", "state-versions", "https://state"),
			},
			filterFile:        "./testdata/filterConfig.json",
			shouldErr:         true,
			errValidationFunc: func(err error) bool { return errors.Is(err, tfdrerrors.ErrDestinationNotEmpty{}) },
			errMessage:        "Test copy error when non empty destination state failed",
		},
		{
			origwks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CsvResponder: httpmock.NewStringResponder(404, ""),
			},
			newwks: &testutils.TfeTestWks{
				Name: "test2",
			},
			filterFile:        "",
			shouldErr:         true,
			errValidationFunc: func(err error) bool { return errors.Is(err, tfdrerrors.ErrSourceIsEmpty{}) },
			errMessage:        "Test copy error when source current state not found failed",
		},
		{
			origwks: &testutils.TfeTestWks{
				Name:   "testNoWorkspace",
				Exists: false,
			},
			newwks: &testutils.TfeTestWks{
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
			origwks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutils.NewState(),
				CsvResponder: testutils.NewResponder("test", "state-versions", "https://state"),
			},
			newwks: &testutils.TfeTestWks{
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
			origwks: &testutils.TfeTestWks{
				Name:         "test1",
				Exists:       true,
				CurrentState: testutils.NewState(),
				CsvResponder: testutils.NewResponder("test", "state-versions", "https://state"),
			},
			newwks: &testutils.TfeTestWks{
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
		err := testutils.SetupWksMockHTTPResponses(c.origwks)
		s.NoError(err, c.errMessage)
		err = testutils.SetupWksMockHTTPResponses(c.newwks)
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

func TestCopySuite(t *testing.T) {
	suite.Run(t, new(CopySuite))
}
