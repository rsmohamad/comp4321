package retrieval

import (
	"bytes"
	"comp4321/database"
	"sort"
)

func intersect(list1, list2 [][]byte) (answer [][]byte) {
	i := 0
	j := 0

	for i != len(list1) && j != len(list2) {
		if bytes.Equal(list1[i], list2[j]) {
			answer = append(answer, list1[i])
			i++
			j++
		} else if bytes.Compare(list1[i], list2[j]) == -1 {
			j++
		} else {
			i++
		}
	}

	return
}

func booleanFilter(query []string, viewer *database.Viewer) (docIDs [][]byte) {
	if len(query) == 0 {
		return
	}

	wordDoc := make(map[string][][]byte)

	for _, word := range query {
		wordDoc[word] = viewer.GetContainingPages(word)
	}

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

