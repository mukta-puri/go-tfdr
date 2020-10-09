package filter

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tj/assert"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/config"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/models"
	"github.com/tyler-technologies/go-terraform-state-copy/internal/testutil"
)

type TestSuite struct {
	suite.Suite
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestReadFiltersFromFile() {
	filterConfig, err := readFiltersFromFile("./testdata/filterConfig.json")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), filterConfig)
	assert.Equal(s.T(), 2, len(filterConfig.Filters))
}

func (s *TestSuite) TestCopyStateFilter() {
	var res = testutil.NewStateResources()
	numFilters := 2

	fr, err := StateFilter(res, CopyResourceFilterFunc, "./testdata/filterConfig.json")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), fr)
	assert.Equal(s.T(), numFilters+len(config.GlobalResources), len(fr))
	assert.True(s.T(), contains(fr, "module.test_module_1", "managed", "type_1"))
	assert.True(s.T(), contains(fr, "module.test_module_2", "managed", "type_2"))
	assert.Equal(s.T(), "new_name_1", get(fr, "module.test_module_1", "managed", "type_1").Name)
	assert.Equal(s.T(), "orig_name_2", get(fr, "module.test_module_2", "managed", "type_2").Name)
	assert.Equal(s.T(), "", get(fr, "module.test_module_2", "managed", "type_2").Instances[0].Attributes["attr2"])
}

func (s *TestSuite) TestDeleteStateFilter() {
	var res = testutil.NewStateResources()
	numFilters := 2

	fr, err := StateFilter(res, DeleteResourceFilterFunc, "./testdata/filterConfig.json")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), fr)
	assert.Equal(s.T(), len(res)-len(config.GlobalResources)-numFilters, len(fr))
	assert.False(s.T(), contains(fr, "module.test_module_1", "managed", "type_1"))
	assert.False(s.T(), contains(fr, "module.test_module_2", "managed", "type_2"))
}

func get(s []models.Resource, module string, mode string, typ string) models.Resource {
	for _, a := range s {
		if a.Module == module && a.Mode == mode && a.Type == typ {
			return a
		}
	}
	return models.Resource{}
}

func contains(s []models.Resource, module string, mode string, typ string) bool {
	for _, a := range s {
		if a.Module == module && a.Mode == mode && a.Type == typ {
			return true
		}
	}
	return false
}
