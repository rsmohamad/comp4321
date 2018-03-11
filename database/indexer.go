package database

import (
	"encoding/json"
	"comp4321/models"
	"github.com/boltdb/bolt"
	"sync"
)

type Indexer struct {
	db *bolt.DB
}

// Return an Indexer object from .db file
func LoadIndexer(filename string) (*Indexer, error) {
	var indexer Indexer
	var err error
	indexer.db, err = bolt.Open(filename, 0666, nil)
	if err != nil {
		return nil, err
	}

	// Ensure that all buckets exist
	indexer.db.Update(func(tx *bolt.Tx) error {
		for i := 0; i < NumTable; i++ {
			tx.CreateBucketIfNotExists(intToByte(i))
		}
		return nil
	})
	return &indexer, nil
}

// Drop all tables in database
func (i *Indexer) DropAll() {
	i.db.Update(func(tx *bolt.Tx) error {
		for i := 0; i < NumTable; i++ {
			tx.DeleteBucket(intToByte(i))
			tx.CreateBucket(intToByte(i))
		}
		return nil
	})
}

// Generic id retriever from mapping table
// Forward map table converts textual representation -> unique Id
// Inverse map table converts unique Id -> textual representation
func (i *Indexer) getId(text string, fwMapTable int, invMapTable int) (id []byte) {
	id = nil
	fw := intToByte(fwMapTable)
	inv := intToByte(invMapTable)

	i.db.View(func(tx *bolt.Tx) error {
		forwardMap := tx.Bucket(intToByte(fwMapTable))
		res := forwardMap.Get([]byte(text))
		if res != nil {
			id = make([]byte, len(res))
			copy(id, res)
		}
		return nil
	})

	if id == nil {
		i.db.Batch(func(tx *bolt.Tx) error {
			forwardMap := tx.Bucket(fw)
			uniqueId, _ := forwardMap.NextSequence()
			id = uint64ToByte(uniqueId)
			forwardMap.Put([]byte(text), id)

			invMap := tx.Bucket(inv)
			invMap.Put(id, []byte(text))

			return nil
		})
	}
	return
}

// Get the pageId for the given URL, create new one if does not exist
func (i *Indexer) getPageId(url string) []byte {
	return i.getId(url, UrlToPageId, PageIdToUrl)
}

// Get the wordId for the given word, create new one if does not exist
func (i *Indexer) getWordId(word string) []byte {
	return i.getId(word, WordToWordId, WordIdToWord)
}

func (i *Indexer) updateInverted(word string, pageId []byte, tablename int) {
	wordId := i.getWordId(word)
	i.db.Batch(func(tx *bolt.Tx) error {
		inverted := tx.Bucket(intToByte(tablename))
		wordSet, _ := inverted.CreateBucketIfNotExists(wordId)
		wordSet.Put(pageId, []byte{1})
		return nil
	})
}

func (i *Indexer) updateForward(word string, pageId []byte, tf int, tablename int) {
	wordId := i.getWordId(word)
	i.db.Batch(func(tx *bolt.Tx) error {
		fw := tx.Bucket(intToByte(tablename))
		set, _ := fw.CreateBucketIfNotExists(pageId)
		set.Put(wordId, intToByte(tf))
		return nil
	})
}

// Check if the URL is present in the database
func (i *Indexer) HasUrl(url string) (present bool) {
	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(intToByte(UrlToPageId))
		val := b.Get([]byte(url))
		present = (val != nil)
		return nil
	})
	return
}

// Insert page into the database.
// This will update all mapping tables and indexes.
func (i *Indexer) UpdateOrAddPage(p *models.Document) {
	pageId := i.getPageId(p.Uri)
	var wg sync.WaitGroup

	wg.Add(len(p.Words))
	for word := range p.Words {
		go func(word string) {
			i.updateInverted(word, pageId, InvertedTable)
			i.updateForward(word, pageId, p.Words[word], ForwardTable)
			wg.Done()
		}(word)
	}
	wg.Wait();
	wg.Add(len(p.Titles))
	for word := range p.Titles {
		go func(word string) {
			i.updateInverted(word, pageId, InvertedTableTitle)
			i.updateForward(word, pageId, p.Titles[word], ForwardTableTitle)
			wg.Done()
		}(word)
	}
	wg.Wait()
	i.db.Batch(func(tx *bolt.Tx) error {
		documents := tx.Bucket(intToByte(PageInfo))
		encoded, _ := json.Marshal(p)
		documents.Put(pageId, encoded)
		return nil
	})
}

// TODO
// Update adj list structure
func (i *Indexer) UpdateAdjList() {

}

// TODO
// Update term weights
func (i *Indexer) UpdateTermWeights() {

}

// TODO
// Update page rank
func (i * Indexer) UpdatePageRank() {

}

func (i *Indexer) Close() {
	i.db.Close()
}
