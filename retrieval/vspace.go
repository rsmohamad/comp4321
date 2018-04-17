package retrieval

import (
	"comp4321/database"
	"math"
	"sort"
)

type CosSimResult struct {
	score float64
	docId uint64
}

func cosSim(query []string, docId uint64, viewer *database.Viewer, res *chan *CosSimResult) *CosSimResult {
	var textInnerProduct float64 = 0
	var bodyInnerProduct float64 = 0
	queryMag := math.Sqrt(float64(len(query)))
	docMag := viewer.GetDocumentMagnitude(docId)
	titleMag := viewer.GetTitleMagnitude(docId)

	for _, word := range query {
		textInnerProduct += viewer.GetTfIdf(docId, word)
		bodyInnerProduct += viewer.GetTitleScore(docId, word)
	}

	textScore := textInnerProduct / (queryMag * docMag)
	titleScore := bodyInnerProduct / (queryMag * titleMag)
	score := textScore + titleScore
	rv := &CosSimResult{score, docId}
	*res <- rv
	return rv
}

func getDocumentScores(query []string, viewer *database.Viewer, docsToSearch []uint64) (map[uint64]float64, []uint64) {
	documentScores := make(map[uint64]float64)
	documentIds := make([]uint64, 0)
	res := make(chan *CosSimResult)
	defer close(res)

	for _, id := range docsToSearch {
		_, exist := documentScores[id]
		if !exist {
			go cosSim(query, id, viewer, &res)
			documentScores[id] = 0
			documentIds = append(documentIds, id)
		}
	}

	for range documentIds {
		result := <-res
		documentScores[result.docId] = result.score
	}

	sort.Slice(documentIds, func(i, j int) bool {
		return documentScores[documentIds[i]] > documentScores[documentIds[j]]
	})

	return documentScores, documentIds
}

func vspaceRetrieval(query []string, viewer *database.Viewer) (map[uint64]float64, []uint64) {
	docsToSearch := make([]uint64, 0)

	for _, word := range query {
		docsToSearch = append(docsToSearch, booleanFilter([]string{word}, viewer)...)
	}

	scores, ids := getDocumentScores(query, viewer, docsToSearch)

	upper := int(math.Min(50.0, float64(len(ids))))
	return scores, ids[0:upper]
}
