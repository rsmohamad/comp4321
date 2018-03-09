package models

import (
	"code.cloudfoundry.org/bytefmt"
	"time"
)

type Document struct {
	Title   string
	Uri     string
	Links   []string
	Words   map[string]int
	Len     int
	Modtime int64
}

func (d Document) GetSizeStr() string {
	if d.Len == 0 {
		return "Not available"
	}

	return bytefmt.ByteSize(uint64(d.Len))
}

func (d Document) GetTimeStr() string {
	if d.Modtime < 0 {
		return "No date available"
	}

	t := time.Unix(d.Modtime, 0)
	return t.Format("02 Jan 2006")
}
