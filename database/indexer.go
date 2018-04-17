package database

import (
	"comp4321/models"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
)

// Class for inserting webpages into the db.
// Reads the .db file in read-write mode.
// Only one instance per file can be created.
type Indexer struct {
	db *bolt.DB

	// Temporarily hold inverted index in memory
	wordInverted, titleInverted map[uint64]map[uint64][]int
	wordIdList, titleIdList     []uint64
	mapLock                     sync.Mutex
	idLock                      sync.Mutex
}

// Return an Indexer object from .db file
func LoadIndexer(filename string) (*Indexer, error) {
	var indexer Indexer
	var err error
	indexer.wordInverted = make(map[uint64]map[uint64][]int)
	indexer.titleInverted = make(map[uint64]map[uint64][]int)
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
			forwardMap := tx.Bucket(intToByte(fwMapTable))
			res := forwardMap.Get([]byte(text))
			if res != nil {
				id = make([]byte, len(res))
				copy(id, res)
				return nil
			}
			uniqueId, _ := forwardMap.NextSequence()
			id = uint64ToByte(uniqueId)
			forwardMap.Put([]byte(text), id)
			invMap := tx.Bucket(intToByte(invMapTable))
			invMap.Put(id, []byte(text))
			return nil
		})
	}
	return
}

// Get the pageId for the given URL, create new one if does not exist
func (i *Indexer) getOrCreatePageId(url string) (pageId []byte) {
	i.idLock.Lock()
	defer i.idLock.Unlock()
	pageId = i.getId(url, UrlToPageId, PageIdToUrl)
	i.db.Update(func(tx *bolt.Tx) error {
		fw := tx.Bucket(intToByte(ForwardTable))
		fwTitle := tx.Bucket(intToByte(ForwardTableTitle))
		fw.CreateBucketIfNotExists(pageId)
		fwTitle.CreateBucketIfNotExists(pageId)
		return nil
	})
	return
}

// Get the wordId for the given word, create new one if does not exist
func (i *Indexer) getOrCreateWordId(word string) (rv []byte) {
	rv = i.getId(word, WordToWordId, WordIdToWord)
	return
}

// Update the in-memory inverted index
func (i *Indexer) updateInverted(word string, indexes []int, pageId []byte, isTitle bool) {
	wordId := i.getOrCreateWordId(word)
	wordIdUint64 := byteToUint64(wordId)
	pageIdUint64 := byteToUint64(pageId)
	var postingList map[uint64][]int

	// Critical section - access shared map and slice
	i.mapLock.Lock()
	if !isTitle {
		postingList = i.wordInverted[wordIdUint64]
		if postingList == nil {
			postingList = make(map[uint64][]int)
			i.wordIdList = append(i.wordIdList, wordIdUint64)
		}
		postingList[pageIdUint64] = indexes
		i.wordInverted[wordIdUint64] = postingList
	} else {
		postingList = i.titleInverted[wordIdUint64]
		if postingList == nil {
			postingList = make(map[uint64][]int)
			i.titleIdList = append(i.titleIdList, wordIdUint64)
		}
		postingList[pageIdUint64] = indexes
		i.titleInverted[wordIdUint64] = postingList
	}
	i.mapLock.Unlock()
	// Non critical section
}

func mergeId(id uint64, tablename int, i *Indexer, wg *sync.WaitGroup) {
	i.db.Batch(func(tx *bolt.Tx) error {
		idBytes := uint64ToByte(id)
		inverted := tx.Bucket(intToByte(tablename))
		wordSet, _ := inverted.CreateBucketIfNotExists(idBytes)
		postingList := i.wordInverted[id]
		for docId, idx := range postingList {
			pos := strings.Trim(strings.Replace(fmt.Sprint(idx), " ", ",", -1), "[]")
			wordSet.Put(uint64ToByte(docId), []byte(pos))
		}
		return nil
	})
	wg.Done()
}

// Sort and write the in-memory inverted index to file
func (i *Indexer) FlushInverted() {
	wordIdList := i.wordIdList
	titleIdList := i.titleIdList

	// Sort slices for sequential writes
	sort.Slice(wordIdList, func(i, j int) bool {
		return wordIdList[i] < wordIdList[j]
	})
	sort.Slice(titleIdList, func(i, j int) bool {
		return titleIdList[i] < titleIdList[j]
	})

	wg := sync.WaitGroup{}
	wg.Add(len(wordIdList) + len(titleIdList))
	for _, id := range wordIdList {
		go mergeId(id, InvertedTable, i, &wg)
	}
	for _, id := range titleIdList {
		go mergeId(id, InvertedTableTitle, i, &wg)
	}
	wg.Wait()
}

func (i *Indexer) updateForward(word string, pageId []byte, tf int, tablename int) {
	wordId := i.getOrCreateWordId(word)
	i.db.Batch(func(tx *bolt.Tx) error {
		fw := tx.Bucket(intToByte(tablename))
		set := fw.Bucket(pageId)
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
		maxTfTable := tx.Bucket(intToByte(MaxTf))
		maxTfTable.Put(pageId, intToByte(maxTf))
		return nil
	})
}

func (i *Indexer) getMaxTf(pageId []byte) (maxTf int) {
	i.db.View(func(tx *bolt.Tx) error {
		maxTfTable := tx.Bucket(intToByte(MaxTf))
		maxTf = byteToInt(maxTfTable.Get(pageId))
		return nil
	})
	return
}

// Insert page into the database.
// This will update all mapping tables and indexes.
func (i *Indexer) UpdateOrAddPage(p *models.Document) {
	pageId := i.getOrCreatePageId(p.Uri)
	var wg sync.WaitGroup
	fmt.Println(pageId, p.Uri)

	wg.Add(len(p.Words))
	for word, wordModel := range p.Words {
		go func(w string, t int, idx []int) {
			i.updateInverted(w, idx, pageId, false)
			i.updateForward(w, pageId, t, ForwardTable)
			wg.Done()
		}(word, wordModel.Tf, wordModel.Positions)
	}
	wg.Wait()
	wg.Add(len(p.Titles))
	for word, wordModel := range p.Titles {
		go func(w string, t int, idx []int) {
			i.updateInverted(w, idx, pageId, true)
			i.updateForward(w, pageId, t, ForwardTableTitle)
			wg.Done()
		}(word, wordModel.Tf, wordModel.Positions)
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
	var childIds []uint64
	adjList := make(map[uint64]map[uint64]int)
	i.db.View(func(tx *bolt.Tx) error {
		piBucket := tx.Bucket(intToByte(PageInfo))
		upBucket := tx.Bucket(intToByte(UrlToPageId))

		piBucket.ForEach(func(parentId, decoded []byte) error {
			var p models.Document
			parentIdUint64 := byteToUint64(parentId)
			json.Unmarshal(decoded, &p)
			// Iterate through each link, clean them, and put according to id 1-30.
			for _, el := range p.Links {
				childId := upBucket.Get([]byte(el)) //childId
				if childId == nil {
					continue
				}

				childIdUint64 := byteToUint64(childId)
				parentList := adjList[childIdUint64]
				if parentList == nil {
					parentList = make(map[uint64]int)
					childIds = append(childIds, childIdUint64)
				}
				parentList[parentIdUint64] = len(p.Links)
				adjList[childIdUint64] = parentList
			}
			return nil
		})
		return nil
	})

	sort.Slice(childIds, func(i, j int) bool {
		return childIds[i] < childIds[j]
	})

	i.db.Update(func(tx *bolt.Tx) error {
		alBucket := tx.Bucket(intToByte(AdjList))
		for _, id := range childIds {
			idBytes := uint64ToByte(id)
			pageSet, _ := alBucket.CreateBucketIfNotExists(idBytes)
			for pageId, len := range adjList[id] {
				pageSet.Put(uint64ToByte(pageId), intToByte(len))
			}
		}
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
		pageMag := tx.Bucket(intToByte(PageMagnitude))
		N := float64(ftBucket.Stats().KeyN)

		// Forward Table (PageId - Terms)
		ftBucket.ForEach(func(pageId, val []byte) error {
			words := ftBucket.Bucket(pageId)
			pageSet, _ := twBucket.CreateBucketIfNotExists(pageId)
			maxTf := float64(i.getMaxTf(pageId))

			// Words Bucket (Words - TF)
			words.ForEach(func(wordId, tfByte []byte) error {
				// TF-IDF COMPUTATION
				df := float64(itBucket.Bucket(wordId).Stats().KeyN)
				tf := float64(byteToInt(tfByte))
				tw := tf * math.Log2(N/df) / maxTf
				if wordId != nil {
					pageSet.Put(wordId, float64ToByte(tw))
				}
				return nil
			})

			// Calculate vector magnitude
			var magnitude float64 = 0
			pageSet.ForEach(func(wordId, scoreByte []byte) error {
				score := byteToFloat64(scoreByte)
				magnitude += score * score
				return nil
			})
			magnitude = math.Sqrt(magnitude)
			pageMag.Put(pageId, float64ToByte(magnitude))

			return nil
		})
		return nil
	})
	return
}

// Calculates the PageRank iteratively.
func (i *Indexer) UpdatePageRank() {
	i.db.Update(func(tx *bolt.Tx) error {
		prBucket := tx.Bucket(intToByte(PageRank))
		adjBucket := tx.Bucket(intToByte(AdjList))

		adjBucket.ForEach(func(childID, _ []byte) error {
			prBucket.Put(childID, float64ToByte(1.0))
			return nil
		})

		pageRank := make(map[uint64]float64)

		for i := 0; i < 15; i++ {
			adjBucket.ForEach(func(childID, _ []byte) error {
				parents := adjBucket.Bucket(childID)
				d := 0.15
				totalParentPR := 0.0

				parents.ForEach(func(parentID, totalChild []byte) error {
					parentPR := prBucket.Get(parentID)

					if len(parentPR) > 0 {
						totalParentPR = totalParentPR + (byteToFloat64(parentPR) / float64(byteToInt(totalChild)))
					}
					return nil
				})

				pageRank[byteToUint64(childID)] = 1.0 - d + (d * totalParentPR)

				return nil
			})

			for id, pr := range pageRank {
				prBucket.Put(uint64ToByte(id), float64ToByte(pr))
			}
		}

		return nil
	})
	return
}

func (i *Indexer) Close() {
	i.db.Close()
}
