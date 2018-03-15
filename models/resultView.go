package models

type ResultView struct {
	Query        string
	Results      []*Document
	TotalResults int
	PageNum      int
	Pages        []int
	CurrentPage  int
}
