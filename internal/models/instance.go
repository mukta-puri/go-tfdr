package models

type Instance struct {
	IndexKey      interface{}            `json:"index_key"`
	SchemaVersion interface{}            `json:"schema_version"`
	Attributes    map[string]interface{} `json:"attributes"`
	Private       string                 `json:"private"`
	Dependencies  []string               `json:"dependencies"`
}
