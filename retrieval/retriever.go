package retrieval

import (
	"bytes"
	"comp4321/database"
	"comp4321/models"
	"comp4321/stopword"
	"strings"
)

var index *database.Indexer

func preprocessText(words []string) {
	for i, word := range words {
		if stopword.IsStopWord(word) {
			words = append(words[:i], words[i+1:]...)
		}
	}
}

func convertWordsToID(words *[]string) (idArr [][]byte) {
	for _, word := range *words {
		idArr = append(idArr, index.GetWordId(word))
	}
	return
}

func booleanFilter(queryIDs *[][]byte) [][]byte {
	var first, docIDs [][]byte

	for i, queryID := range *queryIDs {
		if i == 0 {
			first = append(first, index.GetDocID(queryID)...)
		} else if i == 1 {
			for _, docID := range index.GetDocID(queryID) {
				found := false
				for _, id := range first {
					if bytes.Compare(docID, id) == 0 {
						found = true
						break
					}
				}
				if found {
					docIDs = append(docIDs, docID)
				}
			}
		} else {
			for _, docID := range index.GetDocID(queryID) {
				found := false
				for _, id := range docIDs {
					if bytes.Compare(docID, id) == 0 {
						found = true
						break
					}
				}
				if found {
					docIDs = append(docIDs, docID)
				}
			}
		}
	}

	return docIDs
}

func searchQuery(queryIDs *[][]byte) (documents []models.Document) {
	docIDs := booleanFilter(queryIDs)

	documents = index.GetDocument(docIDs)

	return
}

// Search returns a list of documents that can be found using the query
func Search(query string) []models.Document {
	index, _ = database.LoadIndexer("index.db")
	defer index.Close()

	querySplit := strings.Split(query, " ")

	preprocessText(querySplit)

	queryID := convertWordsToID(&querySplit)

	results := searchQuery(&queryID)

	return results
}
