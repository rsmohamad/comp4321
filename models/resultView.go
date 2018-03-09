package models

type ResultView struct {
	Query string
	Results []*Document
}