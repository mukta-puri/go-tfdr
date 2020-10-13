package api

import (
	"encoding/json"
	"errors"
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

type UtilTestSuite struct {
	suite.Suite
}

func (s *UtilTestSuite) SetupSuite() {
	os.Setenv("TF_TEAM_TOKEN", "test")
	os.Setenv("TF_ORG_NAME", "team")
	config.InitConfig("")
	logging.InitLogger()
}

func (s *UtilTestSuite) SetupTest() {
	httpmock.ActivateNonDefault(httpClient)
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/ping", httpmock.NewStringResponder(204, ""))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/test", newResponder("test", "workspaces", ""))
}

func (s *UtilTestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
}

func (s *UtilTestSuite) TestCreateTFStateVersion() {
	state := testutil.NewState()
	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test/state-versions", newResponder("test", "state-versions", "https://state"))

	err := createTFStateVersion(state, "test")
	s.NoError(err)
}

func (s *UtilTestSuite) TestCreateTFStateVersionNoWorkspace() {
	state := testutil.NewState()
	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test/state-versions", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/not-found", httpmock.NewStringResponder(404, ""))

	err := createTFStateVersion(state, "not-found")
	s.Error(err)
	s.True(errors.Is(err, tfdrerrors.ErrGetWorkspace{
		Err: tfe.ErrResourceNotFound,
	}))
	httpmock.DeactivateAndReset()
}

func (s *UtilTestSuite) TestPullTFState() {
	currentState, err := json.Marshal(testutil.NewState())
	s.NoError(err)

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, string(currentState)))
	st, err := pullTFState("test")
	s.NoError(err)
	s.NotNil(st)
	s.Equal(testutil.DefaultNumResources(), len(st.Resources))
}

func (s *UtilTestSuite) TestPullTFStateNoState() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", httpmock.NewStringResponder(404, ""))
	st, err := pullTFState("test")
	s.NoError(err)
	s.Nil(st)
}

func (s *UtilTestSuite) TestPullTFStateErrGetCurrentState() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", httpmock.NewErrorResponder(errors.New("Error getting current state version")))

	st, err := pullTFState("test")
	s.Error(err)
	s.True(strings.Contains(err.Error(), "Cannot get current state. Error:"))
	s.Nil(st)
}

func (s *UtilTestSuite) TestPullTFStateErrDownloadState() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewErrorResponder(errors.New("Error downloading state")))
	st, err := pullTFState("test")
	s.Error(err)
	s.True(strings.Contains(err.Error(), "Cannot download state. Error:"))
	s.Nil(st)
}
