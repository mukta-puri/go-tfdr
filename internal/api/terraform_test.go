package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
	g "github.com/tyler-technologies/go-terraform-state-copy/internal/config/globalresources"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/logging"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/testutil"
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
	origState, err := json.Marshal(testutil.NewState())
	assert.NoError(s.T(), err)

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, string(origState)))

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/test2", newResponder("test2", "workspaces", ""))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test2/current-state-version", httpmock.NewStringResponder(404, ""))

	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test2/state-versions", func(req *http.Request) (*http.Response, error) {
		state, err := decodeStateFromBody(req)
		assert.NoError(s.T(), err)

		numFilters := 2

		assert.Equal(s.T(), testutil.DefaultTerraformVersion, state.TerraformVersion)
		assert.Equal(s.T(), testutil.DefaultVersion, state.Version)
		assert.Equal(s.T(), "", state.Lineage)
		assert.Equal(s.T(), int64(1), state.Serial)
		assert.Equal(s.T(), numFilters+len(g.GlobalResources), len(state.Resources))

		resp, err := newJSONResponse("test2", "state-versions", "https://state")
		assert.NoError(s.T(), err)

		return resp, nil
	})

	err = CopyTFState("test", "test2", "./testdata/filterConfig.json")
	assert.NoError(s.T(), err)
}

func (s *TestSuite) TestDeleteTFStateResources() {
	origState, err := json.Marshal(testutil.NewState())
	assert.NoError(s.T(), err)

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, string(origState)))

	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test/state-versions", func(req *http.Request) (*http.Response, error) {
		state, err := decodeStateFromBody(req)
		assert.NoError(s.T(), err)

		numFilters := 2

		assert.Equal(s.T(), testutil.DefaultTerraformVersion, state.TerraformVersion)
		assert.Equal(s.T(), testutil.DefaultVersion, state.Version)
		assert.Equal(s.T(), testutil.DefaultLineage, state.Lineage)
		assert.Equal(s.T(), testutil.DefaultSerial+1, state.Serial)
		assert.Equal(s.T(), testutil.DefaultNumResources()-numFilters-len(g.GlobalResources), len(state.Resources))

		resp, err := newJSONResponse("test", "state-versions", "https://state")
		assert.NoError(s.T(), err)

		return resp, nil
	})

	err = DeleteTFStateResources("test", "./testdata/filterConfig.json")
	assert.NoError(s.T(), err)
}

func (s *TestSuite) TestCreateTFStateVersion() {

	httpmock.RegisterResponder("POST", "https://app.terraform.io/api/v2/workspaces/test/state-versions", newResponder("test", "state-versions", "https://state"))

	var state models.State

	err := json.Unmarshal(readFile(s.T(), "./testdata/state.json"), &state)
	assert.NoError(s.T(), err)

	err = createTFStateVersion(state, "test")
	assert.NoError(s.T(), err)

}

func (s *TestSuite) TestPullTFStateNoState() {
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", httpmock.NewStringResponder(404, ""))
	st, err := pullTFState("test")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), st)
	assert.Equal(s.T(), 0, st.Version)
	assert.Equal(s.T(), "", st.TerraformVersion)
	assert.Equal(s.T(), "", st.Lineage)
	assert.Equal(s.T(), int64(0), st.Serial)
	assert.Equal(s.T(), 0, len(st.Resources))
	assert.Nil(s.T(), st.Outputs)
}

func (s *TestSuite) TestPullTFState() {
	var currentState = string(readFile(s.T(), "./testdata/state.json"))

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, currentState))
	st, err := pullTFState("test")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), st)
	assert.Equal(s.T(), 1, len(st.Resources))
	os.Unsetenv("TF_TEAM_TOKEN")
	os.Unsetenv("TF_ORG_NAME")
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
