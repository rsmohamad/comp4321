package database

import (
	"encoding/binary"
	"encoding/json"
	"models"
	"sync"

	"github.com/boltdb/bolt"
)

const WordIdTable = "wordIDs"
const PageIdTable = "pageIds"
const ForwardTable = "forwardIndex"
const InvertedTable = "invertedIndex"

var TableNames = [4]string{WordIdTable, PageIdTable, ForwardTable, InvertedTable}

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

// Get the pageId for the given URL, create new one if does not exist
func (i *Indexer) getPageId(url string) (id []byte) {
	i.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(PageIdTable))
		res := b.Get([]byte(url))
		if res != nil {
			id = make([]byte, len(res))
			copy(id, res)
			return nil
		}
		pageId, _ := b.NextSequence()
		id = itob(pageId)
		b.Put([]byte(url), id)
		return nil
	})
	return
}

// Get the wordId for the given URL, create new one if does not exist
func (i *Indexer) getWordId(word string) (id []byte) {
	i.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(WordIdTable))
		res := b.Get([]byte(word))
		if res != nil {
			id = make([]byte, len(res))
			copy(id, res)
			return nil
		}
		wordId, _ := b.NextSequence()
		id = itob(wordId)
		b.Put([]byte(word), id)
		return nil
	})
	return
}

func (i *Indexer) UpdateOrAddPage(p *models.Document) {
	pageId := i.getPageId(p.Uri)
	addWord := func(word string, wg *sync.WaitGroup) {
		i.updateInverted(word, pageId)
		wg.Done()
	}

	// Update words in
	// Use goroutine to do checking concurrently
	var wg sync.WaitGroup
	wg.Add(len(p.Words))
	for word := range p.Words {
		go addWord(word, &wg)
	}
	wg.Wait()

	i.db.Batch(func(tx *bolt.Tx) error {
		documents := tx.Bucket([]byte(ForwardTable))
		encoded, _ := json.Marshal(p)
		documents.Put(pageId, encoded)
		return nil
	})
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

func (i *Indexer) ForEachWord(fn func(word string, i int)) {
	i.db.View(func(tx *bolt.Tx) error {
		// Get inverted index
		inverted := tx.Bucket([]byte(InvertedTable))

		inverted.ForEach(func(k, v []byte) error {
			return nil
		})

		return nil
	})
}

func (i *Indexer) Close() {
	i.db.Close()
}
