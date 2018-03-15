package retrieval

import (
	"bytes"
	"comp4321/database"
	"comp4321/models"
	"comp4321/stopword"
	"fmt"
	"sort"
	"strings"

	"github.com/surgebase/porter2"
)

var viewer *database.Viewer

func preprocessText(words []string) (rv []string) {
	for _, word := range words {
		if !stopword.IsStopWord(word) {
			cleaned := strings.TrimSpace(word)
			cleaned = strings.Replace(cleaned, "\n", "", -1)
			cleaned = strings.Replace(cleaned, "\r", "", -1)
			cleaned = porter2.Stem(cleaned)
			rv = append(rv, cleaned)
		}
	}

	return
}

func intersect(list1, list2 [][]byte) (answer [][]byte) {
	i := 0
	j := 0

	for i != len(list1) && j != len(list2) {
		if bytes.Equal(list1[i], list2[j]) {
			answer = append(answer, list1[i])
			i++
			j++
		} else if bytes.Compare(list1[i], list2[j]) == -1 {
			i++
		} else {
			j++
		}
	}

	return
}

func booleanFilter(query []string) (docIDs [][]byte) {
	wordDoc := make(map[string][][]byte)

	for _, word := range query {
		wordDoc[word] = viewer.GetContainingPages(word)
	}

	fmt.Println(wordDoc)

	sort.Slice(query, func(i, j int) bool {
		return len(wordDoc[query[i]]) < len(wordDoc[query[j]])
	})

	docIDs = append(docIDs, wordDoc[query[0]]...)

	for i, word := range query {
		if i == 0 {
			continue
		}

		docIDs = intersect(docIDs, wordDoc[word])
	}

	return
}

func searchQuery(query []string) (documents []*models.Document) {
	docIDs := booleanFilter(query)

	documents = viewer.GetDocuments(docIDs)

	return
}

// Search returns a list of documents that can be found using the query
func Search(query string) []*models.Document {
	viewer, _ = database.LoadViewer("index.db")
	defer viewer.Close()

	querySplit := strings.Split(query, " ")

	preprocessed := preprocessText(querySplit)

	results := searchQuery(preprocessed)

	return results
}
