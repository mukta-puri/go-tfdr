package filter

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tyler-technologies/go-tfdr/internal/models"
	"github.com/tyler-technologies/go-tfdr/internal/testutils"
)

type TestSuite struct {
	suite.Suite
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestReadFiltersFromFile() {
	filterConfig, err := readFiltersFromFile("./testdata/filterConfig.json")
	s.NoError(err)
	s.NotNil(filterConfig)
	s.Equal(2, len(filterConfig.Filters))
}

func (s *TestSuite) TestReadFiltersFromFileError() {
	filterConfig, err := readFiltersFromFile("./testdata/not-found.json")
	s.Error(err)
	s.Nil(filterConfig)
}

func (s *TestSuite) TestReadStateFilterError() {
	var res = testutils.NewStateResources()

	fr, err := StateFilter(res, CopyResourceFilterFunc, "./testdata/not-found.json")
	s.Error(err)
	s.Nil(fr)
}

func (s *TestSuite) TestCopyStateFilter() {
	var res = testutils.NewStateResources()
	numFilters := 2

	fr, err := StateFilter(res, CopyResourceFilterFunc, "./testdata/filterConfig.json")
	s.NoError(err)
	s.NotNil(fr)
	s.Equal(numFilters+len(testutils.GlobalResources), len(fr))
	s.True(contains(fr, "module.test_module_1", "managed", "type_1"))
	s.True(contains(fr, "module.test_module_2", "managed", "type_2"))
	s.Equal("new_name_1", get(fr, "module.test_module_1", "managed", "type_1").Name)
	s.Equal("orig_name_2", get(fr, "module.test_module_2", "managed", "type_2").Name)
	s.Equal("", get(fr, "module.test_module_2", "managed", "type_2").Instances[0].Attributes["attr2"])
}

func (s *TestSuite) TestDeleteStateFilter() {
	var res = testutils.NewStateResources()
	numFilters := 2

	fr, err := StateFilter(res, DeleteResourceFilterFunc, "./testdata/filterConfig.json")
	s.NoError(err)
	s.NotNil(fr)
	s.Equal(len(res)-len(testutils.GlobalResources)-numFilters, len(fr))
	s.False(contains(fr, "module.test_module_1", "managed", "type_1"))
	s.False(contains(fr, "module.test_module_2", "managed", "type_2"))
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
