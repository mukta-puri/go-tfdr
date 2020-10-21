package models

type FilterConfig struct {
	GlobalResourceTypes []string `json:"global_resource_types"`
	Filters             []Filter `json:"filters"`
}
