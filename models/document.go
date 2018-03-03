package models

type Document struct {
	Title   string
	Uri     string
	Links   []string
	Words   map[string]int
	Len     int64
	Modtime int64
}
