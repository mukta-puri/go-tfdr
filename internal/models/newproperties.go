package models

type NewProperties struct {
	Name       string                 `json:"name"`
	Attributes map[string]interface{} `json:"attributes"`
}
