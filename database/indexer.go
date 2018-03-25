package database

import (
	"comp4321/models"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"math"
	"net/url"
	"sort"
	"sync"
)

// Class for inserting webpages into the db.
// Reads the .db file in read-write mode.
// Only one instance per file can be created.
type Indexer struct {
	db *bolt.DB

	// Temporarily hold inverted index in memory
	tempInverted map[uint64]map[uint64]bool
	wordIdList   []uint64
	mapLock      sync.Mutex
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

	i.db.Batch(func(tx *bolt.Tx) error {
		forwardMap := tx.Bucket(fw)
		res := forwardMap.Get([]byte(text))

		// Check if the ID already exist
		if res != nil {
			id = make([]byte, len(res))
			copy(id, res)
			return nil
		}
		uniqueId, _ := forwardMap.NextSequence()
		id = uint64ToByte(uniqueId)
		forwardMap.Put([]byte(text), id)
		invMap := tx.Bucket(inv)
		invMap.Put(id, []byte(text))
		return nil
	})

	return
}

// Get the pageId for the given URL, create new one if does not exist
func (i *Indexer) getOrCreatePageId(url string) []byte {
	return i.getId(url, UrlToPageId, PageIdToUrl)
}

// Get the wordId for the given word, create new one if does not exist
func (i *Indexer) getOrCreateWordId(word string) (rv []byte) {
	rv = i.getId(word, WordToWordId, WordIdToWord)
	return
}

// Update the in-memory inverted index
func (i *Indexer) updateInverted(word string, pageId []byte, tablename int) {
	wordId := i.getOrCreateWordId(word)
	wordIdUint64 := byteToUint64(wordId)
	pageIdUint64 := byteToUint64(pageId)

	// Critical section - access shared map and slice
	i.mapLock.Lock()
	if i.tempInverted == nil {
		i.tempInverted = make(map[uint64]map[uint64]bool)
	}

	postingList := i.tempInverted[wordIdUint64]
	if postingList == nil {
		postingList = make(map[uint64]bool)
		i.wordIdList = append(i.wordIdList, wordIdUint64)
	}

	postingList[pageIdUint64] = true
	i.tempInverted[wordIdUint64] = postingList
	i.mapLock.Unlock()
	// Non critical section
}

// Sort and write the in-memory inverted index to file
func (i *Indexer) FlushInverted() {
	wordIdList := i.wordIdList

	// Sort slices for sequential writes
	sort.Slice(wordIdList, func(i, j int) bool {
		return wordIdList[i] < wordIdList[j]
	})

	var wg sync.WaitGroup
	wg.Add(len(wordIdList))
	for index, id := range wordIdList {
		idBytes := uint64ToByte(id)
		fmt.Printf("Merging word %d out of %d | WordID: ", index+1, len(wordIdList))
		fmt.Println(idBytes)
		go i.db.Batch(func(tx *bolt.Tx) error {
			inverted := tx.Bucket(intToByte(InvertedTable))
			wordSet, _ := inverted.CreateBucketIfNotExists(idBytes)
			postingList := i.tempInverted[id]
			for docId, _ := range postingList {
				wordSet.Put(uint64ToByte(docId), []byte{1})
			}

			wg.Done()
			return nil
		})
	}
	wg.Wait()
}

func (i *Indexer) updateForward(word string, pageId []byte, tf int, tablename int) {
	wordId := i.getOrCreateWordId(word)
	i.db.Batch(func(tx *bolt.Tx) error {
		fw := tx.Bucket(intToByte(tablename))
		set, _ := fw.CreateBucketIfNotExists(pageId)
		set.Put(wordId, intToByte(tf))
		return nil
	})
}

// Check if the URL is present in the database
func (i *Indexer) ContainsUrl(url string) (present bool) {
	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(intToByte(UrlToPageId))
		val := b.Get([]byte(url))
		present = val != nil
		return nil
	})
	return
}

func (i *Indexer) setMaxTf(pageId []byte, maxTf int) {
	i.db.Batch(func(tx *bolt.Tx) error {
		fwTable := tx.Bucket(intToByte(ForwardTable))
		fwTable.Put(pageId, intToByte(maxTf))
		return nil
	})
}

func (i *Indexer) getMaxTf(pageId []byte) (maxTf int) {
	i.db.View(func(tx *bolt.Tx) error {
		fwTable := tx.Bucket(intToByte(ForwardTable))
		maxTf = byteToInt(fwTable.Get(pageId))
		return nil
	})
	return
}

// Insert page into the database.
// This will update all mapping tables and indexes.
func (i *Indexer) UpdateOrAddPage(p *models.Document) {
	pageId := i.getOrCreatePageId(p.Uri)
	var wg sync.WaitGroup

	wg.Add(len(p.Words))
	for word, tf := range p.Words {
		go func(w string, t int) {
			i.updateInverted(w, pageId, InvertedTable)
			i.updateForward(w, pageId, t, ForwardTable)
			wg.Done()
		}(word, tf)
	}
	wg.Wait()

	i.setMaxTf(pageId, p.MaxTf)
	i.db.Batch(func(tx *bolt.Tx) error {
		documents := tx.Bucket(intToByte(PageInfo))
		encoded, _ := json.Marshal(p)
		documents.Put(pageId, encoded)
		return nil
	})
}

// Update Adjacency List
// Gets the pageId and set of child Links from PageInfo
// Sets in each of the child link, the pageId as parent link and the number of links from the pageId
func (i *Indexer) UpdateAdjList() {
	i.db.Update(func(tx *bolt.Tx) error {
		piBucket := tx.Bucket(intToByte(PageInfo))
		upBucket := tx.Bucket(intToByte(UrlToPageId))
		alBucket := tx.Bucket(intToByte(AdjList))

		// PageInfo Table (pageId - JSON Document)
		piBucket.ForEach(func(pageId, decoded []byte) error {
			var p models.Document
			json.Unmarshal(decoded, &p)
			Links := p.Links
			// Iterate through each link, clean them, and put according to id 1-30.
			for _, el := range Links {
				u, _ := url.Parse(el)
				newUrl := u.Scheme + "://" + u.Host + u.Path
				id := upBucket.Get([]byte(newUrl)) //childId
				if byteToInt(id) != 0 {
					pageSet, _ := alBucket.CreateBucketIfNotExists(id)
					pageSet.Put(pageId, intToByte(len(Links)))
				}
			}
			return nil
		})

		return nil
	})
}

// Update term weights
// TF, N, keywords per page, and pages are retrieved from forward table
// DF is retrieved from inverted index
func (i *Indexer) UpdateTermWeights() {
	i.db.Update(func(tx *bolt.Tx) error {
		itBucket := tx.Bucket(intToByte(InvertedTable))
		ftBucket := tx.Bucket(intToByte(ForwardTable))
		twBucket := tx.Bucket(intToByte(TermWeights))
		N := float64(ftBucket.Stats().KeyN)

		// Forward Table (PageId - Terms)
		ftBucket.ForEach(func(pageId, _ []byte) error {
			words := ftBucket.Bucket(pageId)
			pageSet, _ := twBucket.CreateBucketIfNotExists(pageId)
			maxTf := float64(i.getMaxTf(pageId))

			// Words Bucket (Words - TF)
			words.ForEach(func(wordId, tfByte []byte) error {
				// TF-IDF COMPUTATION
				df := float64(itBucket.Bucket(wordId).Stats().KeyN)
				tf := float64(byteToInt(tfByte))
				tw := tf * math.Log2(N/df) / maxTf
				pageSet.Put(wordId, float64ToByte(tw))
				return nil
			})
			return nil
		})
		return nil
	})
	return
}

// TODO
// Update page rank
func (i *Indexer) UpdatePageRank() {

}

func (i *Indexer) Close() {
	i.db.Close()
}
