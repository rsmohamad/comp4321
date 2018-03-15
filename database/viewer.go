package database

import (
	"encoding/json"
	"comp4321/models"
	"github.com/boltdb/bolt"
)

// The Viewer object abstracts away data structure manipulations
// for reading a search engine database.
// The Viewer object will read the .db file in read-only mode.
// Multiple Viewer objects can read the same .db file simultaneously.
type Viewer struct {
	db *bolt.DB
}

// Load a Viewer object from .db file
func LoadViewer(filename string) (*Viewer, error) {
	var viewer Viewer
	var err error
	viewer.db, err = bolt.Open(filename, 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	return &viewer, nil
}

func (v *Viewer) containsKey(key string, table int) (present bool) {
	v.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(intToByte(table))
		val := b.Get([]byte(key))
		present = val != nil
		return nil
	})
	return
}

func (v *Viewer) stringToId(key string, table int) (rv []byte) {
	rv = nil
	v.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(intToByte(table))
		val := b.Get([]byte(key))
		if val != nil {
			rv = make([]byte, len(val))
			copy(rv, val)
		}
		return nil
	})
	return
}

func (v *Viewer) wordToId(word string) []byte {
	return v.stringToId(word, WordToWordId)
}

func (v *Viewer) urlToId(url string) []byte {
	return v.stringToId(url, UrlToPageId)
}

// Check if a URL exist in the database
func (v *Viewer) ContainsUrl(url string) bool {
	return v.containsKey(url, UrlToPageId)
}

// Check if a word exist in the database
func (v *Viewer) ContainsWord(word string) bool {
	return v.containsKey(word, WordToWordId)
}

// Returns a list of page IDs containing a word
func (v *Viewer) GetContainingPages(word string) [][]byte {
	rv := make([][]byte, 0)
	wordId := v.wordToId(word)
	if wordId == nil {
		return rv
	}
	v.db.View(func(tx *bolt.Tx) error {
		wordBucket := tx.Bucket(intToByte(InvertedTable))
		docBucket := wordBucket.Bucket(wordId)
		docBucket.ForEach(func(k, v []byte) error {
			pageId := make([]byte, len(k))
			copy(pageId, k)
			rv = append(rv, pageId)
			return nil
		})
		return nil
	})
	return rv
}

// Returns a document object from a pageId.
// Returns nil if the pageId does not exist
func (v *Viewer) GetDocument(pageId []byte) (document *models.Document) {
	v.db.View(func(tx *bolt.Tx) error {
		documents := tx.Bucket(intToByte(PageInfo))
		docBytes := documents.Get(pageId)

		// pageId does not exist if docBytes == nil
		if docBytes == nil {
			return nil
		}

		document = &models.Document{}
		err := json.Unmarshal(docBytes, documents)

		// If there are parsing errors, return nil pointer
		if err != nil {
			document = nil
		}
		return nil
	})
	return
}

// Returns a list of documents from a list of IDs
func (v *Viewer) GetDocuments(pageIds [][]byte) (documents []*models.Document) {
	documents = make([]*models.Document, 0)
	for _, pageId := range pageIds {
		document := v.GetDocument(pageId)
		documents = append(documents, document)
	}
	return
}

// Iterate over all documents
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
