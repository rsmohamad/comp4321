package retrieval

import (
	"comp4321/database"
	"math"
	"sort"
)

func cosSim(query []string, docId uint64, viewer *database.Viewer) float64 {
	var innerProduct float64 = 0
	queryMag := math.Sqrt(float64(len(query)))
	docMag := viewer.GetDocumentMagnitude(docId)

	for _, word := range query {
		innerProduct += viewer.GetTfIdf(docId, word)
	}

	return innerProduct / (queryMag * docMag)
}

func vspaceRetrieval(query []string, viewer *database.Viewer) (map[uint64]float64, []uint64) {
	documentScores := make(map[uint64]float64)
	documentIds := make([]uint64, 0)

	for _, word := range query {
		docsToSearch := booleanFilter([]string{word}, viewer)
		for _, id := range docsToSearch {
			_, exist := documentScores[id]
			if !exist {
				documentScores[id] = cosSim(query, id, viewer)
				documentIds = append(documentIds, id)
			}
		}
	}

	sort.Slice(documentIds, func(i, j int) bool {
		return documentScores[documentIds[i]] > documentScores[documentIds[j]]
	})

	return documentScores, documentIds
}
