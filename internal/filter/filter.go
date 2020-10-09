package filter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	g "github.com/tyler-technologies/go-terraform-state-copy/internal/config/globalresources"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
)

// StateFilter &
func StateFilter(vs []models.Resource, f func(models.Resource, *models.FilterConfig) models.Resource, configFileName string) ([]models.Resource, error) {
	filterConfig, err := readFiltersFromFile(configFileName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get state filters. Err: %v", err)
	}
	vsf := make([]models.Resource, 0)
	for _, v := range vs {
		result := f(v, &filterConfig)
		if result.Module != "" {
			vsf = append(vsf, result)
		}
	}
	return vsf, nil
}

// CopyResourceFilterFunc &
var CopyResourceFilterFunc = func(resource models.Resource, filterConfig *models.FilterConfig) models.Resource {
	for _, globalResource := range g.GlobalResources {
		if resource.Type == globalResource {
			return resource
		}
	}
	for _, filter := range filterConfig.Filters {
		if resource.Mode == "managed" && resource.Module == filter.FilterProperties.Module && resource.Name == filter.FilterProperties.Name && resource.Type == filter.FilterProperties.Type {
			if filter.NewProperties.Name != "" {
				resource.Name = filter.NewProperties.Name
			}

			for k, v := range filter.NewProperties.Attributes {
				resource.Instances[0].Attributes[k] = v
			}

			return resource
		}
	}

	return models.Resource{}
}

// DeleteResourceFilterFunc &
var DeleteResourceFilterFunc = func(resource models.Resource, filterConfig *models.FilterConfig) models.Resource {
	for _, globalResource := range g.GlobalResources {
		if resource.Type == globalResource {
			return models.Resource{}
		}
	}

	for _, filter := range filterConfig.Filters {
		if resource.Mode == "managed" && resource.Module == filter.FilterProperties.Module && resource.Name == filter.FilterProperties.Name && resource.Type == filter.FilterProperties.Type {
			return models.Resource{}
		}
	}

	return resource
}

func readFiltersFromFile(configFileName string) (models.FilterConfig, error) {
	filterConfigFile, err := os.Open(configFileName)
	if err != nil {
		return models.FilterConfig{}, fmt.Errorf("Unable to read file. Err: %v", err)
	}
	configByteValue, _ := ioutil.ReadAll(filterConfigFile)

	var filterConfig models.FilterConfig

	json.Unmarshal(configByteValue, &filterConfig)

	return filterConfig, nil
}
