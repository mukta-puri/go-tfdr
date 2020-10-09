package models

type FilterProperties struct {
	Module string `json:"module"` // Need to be able to search by wildcard
	Type   string `json:"type"`
	Name   string `json:"name"`
}
