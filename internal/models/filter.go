package models

type Filter struct {
	Original FilterProperties `json:"original"`
	New      FilterProperties `json:"new"`
}
