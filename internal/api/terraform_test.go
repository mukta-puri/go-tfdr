package api

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
)

func TestPullTFState(t *testing.T) {
	httpmock.ActivateNonDefault(httpClient)
	os.Setenv("TF_TEAM_TOKEN", "test")
	os.Setenv("TF_ORG_NAME", "team")
	config.InitConfig("")

	var currentState = readFile(t, "./testdata/state.json")

	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/ping", httpmock.NewStringResponder(204, ""))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/organizations/team/workspaces/test", newResponder("test", "workspaces", ""))
	httpmock.RegisterResponder("GET", "https://app.terraform.io/api/v2/workspaces/test/current-state-version", newResponder("test", "state-versions", "https://state"))
	httpmock.RegisterResponder("GET", "https://state", httpmock.NewStringResponder(200, currentState))
	st, err := pullTFState("test")
	assert.NoError(t, err)
	assert.NotNil(t, st)
	assert.Equal(t, 1, len(st.Resources))
	os.Unsetenv("TF_TEAM_TOKEN")
	os.Unsetenv("TF_ORG_NAME")
}

func readFile(t *testing.T, filename string) string {
	file, err := os.Open(filename)
	assert.NoError(t, err)
	defer func() {
		err = file.Close()
		assert.NoError(t, err)
	}()
	b, err := ioutil.ReadAll(file)
	assert.NoError(t, err)
	return string(b)
}

func newResponder(id, typ, hostedStateDownloadUrl string) httpmock.Responder {
	res := response{
		Data: responseData{
			Id:  id,
			Typ: typ,
		},
	}
	if hostedStateDownloadUrl != "" {
		res.Data.Attributes = responseAttr{
			HostedStateDownloadUrl: hostedStateDownloadUrl,
		}
	}

	return httpmock.NewJsonResponderOrPanic(200, res)
}

type response struct {
	Data responseData `json:"data"`
}

type responseData struct {
	Id         string       `json:"id"`
	Typ        string       `json:"type"`
	Attributes responseAttr `json:"attributes,omitempty"`
}

type responseAttr struct {
	HostedStateDownloadUrl string `json:"hosted-state-download-url,omitempty"`
}
