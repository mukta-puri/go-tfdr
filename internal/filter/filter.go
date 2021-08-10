package filter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mupuri/go-tfdr/internal/models"
	"github.com/mupuri/go-tfdr/internal/tfdrerrors"
)

// StateFilter &
func StateFilter(vs []models.Resource, f func(*models.Resource, *models.FilterConfig) *models.Resource, configFileName string) ([]models.Resource, error) {
	filterConfig, err := readFiltersFromFile(configFileName)
	if err != nil {
		return nil, tfdrerrors.ErrReadFilterFile{Err: err}
	}
	vsf := make([]models.Resource, 0)
	for _, v := range vs {
		result := f(&v, filterConfig)
		if result != nil {
			vsf = append(vsf, *result)
		}
	}
	return vsf, nil
}

// CopyResourceFilterFunc &
var CopyResourceFilterFunc = func(resource *models.Resource, filterConfig *models.FilterConfig) *models.Resource {
	for _, globalResource := range filterConfig.GlobalResourceTypes {
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

	return nil
}

// DeleteResourceFilterFunc &
var DeleteResourceFilterFunc = func(resource *models.Resource, filterConfig *models.FilterConfig) *models.Resource {
	for _, globalResource := range filterConfig.GlobalResourceTypes {
		if resource.Type == globalResource {
			return nil
		}
	}

	for _, filter := range filterConfig.Filters {
		if resource.Mode == "managed" && resource.Module == filter.FilterProperties.Module && resource.Name == filter.FilterProperties.Name && resource.Type == filter.FilterProperties.Type {
			return nil
		}
	}

	return resource
}

func readFiltersFromFile(configFileName string) (*models.FilterConfig, error) {
	filterConfigFile, err := os.Open(configFileName)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file. Err: %v", err)
	}
	configByteValue, _ := ioutil.ReadAll(filterConfigFile)

	var filterConfig models.FilterConfig

	json.Unmarshal(configByteValue, &filterConfig)

	return &filterConfig, nil
}
