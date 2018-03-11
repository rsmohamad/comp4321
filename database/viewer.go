package database

import (
	"encoding/json"
	"comp4321/models"
	"github.com/boltdb/bolt"
)

type Viewer struct {
	db *bolt.DB
}

func LoadViewer(filename string) (*Viewer, error) {
	var viewer Viewer
	var err error
	viewer.db, err = bolt.Open(filename, 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	return &viewer, nil
}

// Function for iterating over all documents
func (v *Viewer) ForEachDocument(fn func(p *models.Document, i int)) {
	v.db.View(func(tx *bolt.Tx) error {
		// Get pages bucket
		documents := tx.Bucket(intToByte(PageInfo))
		count := 0

		// Iterate over all documents
		documents.ForEach(func(k, v []byte) error {
			page := &models.Document{}
			err := json.Unmarshal(v, page)
			if err == nil {
				fn(page, count)
				count++
			}
			return nil
		})
		return nil
	})
}

func (v *Viewer) Close() {
	v.db.Close()
}
