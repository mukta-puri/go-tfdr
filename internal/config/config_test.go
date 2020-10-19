package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	vpr "github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
}

func (s *TestSuite) SetupTest() {
	os.Unsetenv("TF_TEAM_TOKEN")
	os.Unsetenv("TF_ORG_NAME")
	os.Unsetenv("TF_STATE_COPY_LOG_LEVEL")
	viper = vpr.New()
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestValidateConfig() {
	cases := []struct {
		tftoken   string
		tforgname string
		errorType error
		message   string
	}{
		{"", "", ErrTFTeamTokenRequired, "empty configuration should return ErrTFTeamTokenRequired"},
		{"", "test", ErrTFTeamTokenRequired, "empty tf token should return ErrTFTeamTokenRequired"},
		{"test", "", ErrTFOrgNameRequired, "empty tf org name should return ErrTFOrgNameRequired"},
		{"test", "test", nil, "valid configuration should not return an error"},
	}

	for _, c := range cases {
		configuration.TerraformTeamToken, configuration.TerraformOrgName =
			c.tftoken, c.tforgname
		err := ValidateConfig()
		if c.errorType != nil {
			s.True(errors.Is(err, c.errorType), c.message)
		} else {
			s.NoError(err, c.message)
		}
	}
}

func createTestFile(filepath, tftoken, tforgname, loglevel string) error {
	content := `tf_team_token: "%s"
tf_org_name: "%s"
tf_state_copy_log_level: "%s"
`

	cfg := []byte(fmt.Sprintf(content, tftoken, tforgname, loglevel))
	return ioutil.WriteFile(filepath, cfg, 0644)
}

func (s *TestSuite) TestReadFromHome() {
	dir := "./test-home"
	os.Setenv("HOME", dir)
	os.MkdirAll(path.Join(dir, ".tfdr"), 0755)
	cfgFile := path.Join(dir, ".tfdr/config.yaml")
	defer os.RemoveAll(dir)
	err := createTestFile(cfgFile, "test_tf_team_token", "test_org_name", "debug")
	s.NoError(err)
	InitConfig("")

	s.Equal("test_tf_team_token", GetConfig().TerraformTeamToken, "tf token should be 'test_tf_team_token'")
	s.Equal("test_org_name", GetConfig().TerraformOrgName, "tf org name should be 'test_org_name'")
	s.Equal("debug", GetConfig().LogLevel, "log level should be 'debug'")
}

func (s *TestSuite) TestInitConfigFile() {
	cfgFile := "./test-config.yml"

	err := createTestFile(cfgFile, "init_tf_team_token", "init_org_name", "debug")
	defer os.RemoveAll(cfgFile)
	s.NoError(err, "should not error creating config file")
	InitConfig(cfgFile)

	s.Equal("init_tf_team_token", configuration.TerraformTeamToken, "tf token should be 'init_tf_team_token'")
	s.Equal("init_org_name", configuration.TerraformOrgName, "tf org name should be 'init_org_name'")
	s.Equal("debug", configuration.LogLevel, "log level should be 'debug'")
}

func (s *TestSuite) TestInitConfigEnv() {
	cfgFile := "./config-env-test.yaml"
	os.Create(cfgFile)
	defer os.RemoveAll(cfgFile)
	os.Setenv("TF_TEAM_TOKEN", "team_token")
	os.Setenv("TF_ORG_NAME", "org_name")
	os.Setenv("TF_STATE_COPY_LOG_LEVEL", "debug")

	InitConfig(cfgFile)

	s.Equal("team_token", configuration.TerraformTeamToken, "tf token should be 'team_token'")
	s.Equal("org_name", configuration.TerraformOrgName, "tf org name should be 'org_name'")
	s.Equal("debug", configuration.LogLevel, "log level should be 'debug'")
}

func (s *TestSuite) TestInitConfigFileOverrides() {
	cfgFile := "./config-override-test.yml"
	os.Setenv("TF_TEAM_TOKEN", "env_team_token")
	os.Setenv("TF_ORG_NAME", "env_org_name")
	os.Setenv("TF_STATE_COPY_LOG_LEVEL", "env_debug")

	err := createTestFile(cfgFile, "overridden_team_token", "overriden_org_name", "info")
	defer os.RemoveAll(cfgFile)
	s.NoError(err, "should not error creating config file")
	InitConfig(cfgFile)

	s.Equal("env_team_token", configuration.TerraformTeamToken, "tf token should be 'env_team_token'")
	s.Equal("env_org_name", configuration.TerraformOrgName, "tf org name should be 'env_org_name'")
	s.Equal("env_debug", configuration.LogLevel, "log level should be 'env_debug'")
}

func (s *TestSuite) TestCreate() {
	dir := "./fake-home"
	os.Setenv("HOME", dir)
	defer os.RemoveAll(dir)
	var in bytes.Buffer
	in.Write([]byte("team_token\norg_name\n"))
	out := readStdOut(func() {
		GenerateConfig(&in)
	})
	cfgFile := path.Join(dir, ".tfdr/config.yaml")
	s.FileExists(cfgFile)
	s.Contains(out, "\nSuccessfully configured terraform disaster recovery cli. Use `tfdr config get` to view your configuration.")
}

func readStdOut(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	stderr := os.Stderr
	logrus.SetOutput(w)
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		logrus.SetOutput(os.Stderr)
	}()
	os.Stdout = w
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, r)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	w.Close()
	return <-out
}
