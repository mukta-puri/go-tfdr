package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	vpr "github.com/ory/viper"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config/file"
	"gopkg.in/yaml.v2"
)

var configuration *Configuration

// GlobalResources &
var GlobalResources []string = []string{
	"aws_cloudfront_distribution",
	"aws_cloudfront_origin_access_identity",
	"aws_iam_access_key",
	"aws_iam_policy_document",
	"aws_iam_policy",
	"aws_iam_role_policy_attachment",
	"aws_iam_role_policy",
	"aws_iam_role",
	"aws_iam_user_policy",
	"aws_iam_user",
	"aws_route53_record",
	"keycloak_openid_client",
	"okta_app_oauth",
	"okta_app_group_assignment",
}

// ErrTFTeamTokenRequired &
var (
	ErrTFTeamTokenRequired = errors.New("Terraform team token is required")
	ErrTFOrgNameRequired   = errors.New("Terraform team token is required")
	viper                  = vpr.New()
)

// Configuration &
type Configuration struct {
	TerraformTeamToken string `mapstructure:"tf_team_token" yaml:"tf_team_token"`
	TerraformOrgName   string `mapstructure:"tf_org_name" yaml:"tf_org_name"`
	LogLevel           string `mapstructure:"tf_state_copy_log_level" yaml:"tf_state_copy_log_level"`
}

// GetConfig &
func GetConfig() *Configuration {
	return configuration
}

// ValidateConfig &
func ValidateConfig() error {
	if len(configuration.TerraformTeamToken) == 0 {
		return ErrTFTeamTokenRequired
	}
	if len(configuration.TerraformOrgName) == 0 {
		return ErrTFOrgNameRequired
	}
	return nil
}

// New &
func New() *Configuration {
	c := Configuration{
		LogLevel: "info",
	}
	return &c
}

// InitConfig &
func InitConfig(cfgFile string) {
	configuration = New()
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME/.tfdr")
		viper.AddConfigPath(".")
	}
	_ = viper.BindEnv("TF_TEAM_TOKEN")
	_ = viper.BindEnv("TF_ORG_NAME")
	_ = viper.BindEnv("TF_STATE_COPY_LOG_LEVEL")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	if err := viper.Unmarshal(&configuration); err != nil {
		log.Fatalf("ERROR: Error reading config: %v", err)
	}
}

// GenerateConfig &
func GenerateConfig(r io.Reader) {
	c := promptConfig(r)
	bytes, _ := yaml.Marshal(c)
	file.Create(string(bytes))
}

func promptConfig(r io.Reader) Configuration {
	reader := bufio.NewReader(r)
	fmt.Println("Enter Terraform team token: ")
	tfToken, _ := reader.ReadString('\n')

	fmt.Println("Enter Terraform org name: ")
	tfOrgName, _ := reader.ReadString('\n')

	configuration := Configuration{
		TerraformTeamToken: strings.TrimSpace(tfToken),
		TerraformOrgName:   strings.TrimSpace(tfOrgName),
	}
	return configuration
}
