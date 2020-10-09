package models

type Filter struct {
	FilterProperties FilterProperties `json:"filter_properties"`
	NewProperties    NewProperties    `json:"new_properties"`
}
