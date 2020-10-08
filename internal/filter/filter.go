package filter

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
)

var globalResources []string = []string{
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

// StateFilter &
func StateFilter(vs []models.Resource, f func(models.Resource, *models.FilterConfig) models.Resource, configFileName string) []models.Resource {
	filterConfig := readFiltersFromFile(configFileName)
	vsf := make([]models.Resource, 0)
	for _, v := range vs {
		result := f(v, &filterConfig)
		if result.Module != "" {
			vsf = append(vsf, result)
		}
	}
	return vsf
}

// CopyResourceFilterFunc &
var CopyResourceFilterFunc = func(resource models.Resource, filterConfig *models.FilterConfig) models.Resource {
	for _, globalResource := range globalResources {
		if resource.Type == globalResource {
			return resource
		}
	}
	for _, filter := range filterConfig.Filters {
		if resource.Mode == "managed" && resource.Module == filter.Original.Module && resource.Name == filter.Original.Name && resource.Type == filter.Original.Type {
			resource.Module = filter.New.Module
			resource.Name = filter.New.Name
			resource.Type = filter.New.Type

			for k, v := range filter.New.Attributes {
				resource.Instances[0].Attributes[k] = v
			}

			return resource
		}
	}

	return models.Resource{}
}

// DeleteResourceFilterFunc &
var DeleteResourceFilterFunc = func(resource models.Resource, filterConfig *models.FilterConfig) models.Resource {
	for _, globalResource := range globalResources {
		if resource.Type == globalResource {
			return models.Resource{}
		}
	}
	for _, filter := range filterConfig.Filters {
		if resource.Mode == "managed" && resource.Module == filter.Original.Module && resource.Name == filter.Original.Name && resource.Type == filter.Original.Type {
			return models.Resource{}
		}
	}

	return resource
}

func readFiltersFromFile(configFileName string) models.FilterConfig {
	filterConfigFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal(err)
	}
	configByteValue, _ := ioutil.ReadAll(filterConfigFile)

	var filterConfig models.FilterConfig

	json.Unmarshal(configByteValue, &filterConfig)
	return filterConfig
}
