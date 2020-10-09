package testutil

import (
	"fmt"

	g "github.com/tyler-technologies/go-terraform-state-copy/internal/config/globalresources"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
)

var defaultNonGlobalResources int = 10

var DefaultTerraformVersion string = "0.13.4"
var DefaultLineage string = "test"
var DefaultVersion int = 4
var DefaultSerial int64 = int64(1)

func NewState() models.State {
	return models.State{
		Version:          DefaultVersion,
		TerraformVersion: DefaultTerraformVersion,
		Serial:           DefaultSerial,
		Lineage:          DefaultLineage,
		Outputs:          nil,
		Resources:        NewStateResources(),
	}
}

func DefaultNumResources() int {
	return defaultNonGlobalResources + len(g.GlobalResources)
}

func NewStateResources() []models.Resource {
	resources := make([]models.Resource, 0)

	for i := 0; i < 10; i++ {
		res := models.Resource{
			Module: fmt.Sprintf("module.test_module_%v", i),
			Mode:   "managed",
			Type:   fmt.Sprintf("type_%v", i),
			Name:   fmt.Sprintf("orig_name_%v", i),
			Instances: []models.Instance{
				{
					Attributes: map[string]interface{}{
						"attr1": "old_value_1",
						"attr2": "old_value_2",
					},
				},
			},
		}
		resources = append(resources, res)
	}

	for i, v := range g.GlobalResources {
		res := models.Resource{
			Module: fmt.Sprintf("module.test_global_module_%v", i),
			Mode:   "managed",
			Type:   v,
			Name:   fmt.Sprintf("global_orig_name_%v", i),
		}
		resources = append(resources, res)
	}

	return resources
}
