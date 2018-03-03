package database

import (
	"encoding/binary"
	"encoding/json"
	"comp4321/models"
	"github.com/boltdb/bolt"
	"sync"
)

const WordToWordId = "wordIDs"
const UrlToPageId = "pageIDs"
const WordIdToWord = "invWordIDs"
const PageIdToUrl = "invPageIDs"
const ForwardTable = "forwardIndex"
const InvertedTable = "invertedIndex"

var TableNames = [6]string{WordToWordId, WordIdToWord, UrlToPageId, PageIdToUrl, ForwardTable, InvertedTable}

type Indexer struct {
	db *bolt.DB
}

// Return an Indexer object from .db file
func LoadIndexer(filename string) (*Indexer, error) {
	var indexer Indexer
	var err error
	indexer.db, err = bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}

	// Ensure that all buckets exist
	indexer.db.Update(func(tx *bolt.Tx) error {
		for _, table := range TableNames {
			tx.CreateBucketIfNotExists([]byte(table))
		}
		return nil
	})

	return &indexer, nil
}

func (i *Indexer) DropAll() {
	i.db.Update(func(tx *bolt.Tx) error {
		for _, table := range TableNames {
			tx.DeleteBucket([]byte(table))
			tx.CreateBucket([]byte(table))
		}
		return nil
	})
}

// Convert uint64 to array of bytes
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	return b
}

// Generic id retriever from mapping table
// Forward map table converts textual representation -> unique Id
// Inverse map table converts unique Id -> textual representation
func (i *Indexer) getId(text string, fwMapTable string, invMapTable string) (id []byte) {
	id = nil
	i.db.View(func(tx *bolt.Tx) error {
		forwardMap := tx.Bucket([]byte(fwMapTable))
		res := forwardMap.Get([]byte(text))
		if res != nil {
			id = make([]byte, len(res))
			copy(id, res)
		}
		return nil
	})

	if id == nil {
		i.db.Batch(func(tx *bolt.Tx) error {
			forwardMap := tx.Bucket([]byte(fwMapTable))
			uniqueId, _ := forwardMap.NextSequence()
			id = itob(uniqueId)
			forwardMap.Put([]byte(text), id)

			invMap := tx.Bucket([]byte(invMapTable))
			invMap.Put(id, []byte(text))

			return nil
		})
	}
	return
}

// Check if the URL is present in the database
func (i *Indexer) IsUrlPresent(url string) (present bool) {
	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UrlToPageId))
		val := b.Get([]byte(url))
		present = (val != nil)
		return nil
	})
	return
}

// Get the pageId for the given URL, create new one if does not exist
func (i *Indexer) getPageId(url string) []byte {
	return i.getId(url, UrlToPageId, PageIdToUrl)
}

// Get the wordId for the given word, create new one if does not exist
func (i *Indexer) getWordId(word string) []byte {
	return i.getId(word, WordToWordId, WordToWordId)
}

func (i *Indexer) UpdateOrAddPage(p *models.Document) {
	pageId := i.getPageId(p.Uri)
	var wg sync.WaitGroup
	addWord := func(word string) {
		i.updateInverted(word, pageId)
		wg.Done()
	}

	// Update inverted table concurrently
	wg.Add(len(p.Words))
	for word := range p.Words {
		go addWord(word)
	}

	i.db.Batch(func(tx *bolt.Tx) error {
		documents := tx.Bucket([]byte(ForwardTable))
		encoded, _ := json.Marshal(p)
		documents.Put(pageId, encoded)
		return nil
	})
	wg.Wait()
}

func (i *Indexer) updateInverted(word string, pageId []byte) {
	// Inverted index consists of <wordId, set>
	wordId := i.getWordId(word)

	i.db.Batch(func(tx *bolt.Tx) error {
		inverted := tx.Bucket([]byte(InvertedTable))
		wordSet, _ := inverted.CreateBucketIfNotExists(wordId)
		wordSet.Put(pageId, []byte{1})
		return nil
	})
}

// Function for iterating over all documents
func (i *Indexer) ForEachDocument(fn func(p *models.Document, i int)) {
	i.db.View(func(tx *bolt.Tx) error {
		// Get pages bucket
		documents := tx.Bucket([]byte(ForwardTable))
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

func (i *Indexer) Close() {
	i.db.Close()
}
