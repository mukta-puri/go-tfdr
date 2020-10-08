package models

type State struct {
	Version          int         `json:"version"`
	TerraformVersion string      `json:"terraform_version"`
	Serial           int64       `json:"serial"`
	Lineage          string      `json:"lineage"`
	Outputs          interface{} `json:"outputs"`
	Resources        []Resource  `json:"resources"`
}
