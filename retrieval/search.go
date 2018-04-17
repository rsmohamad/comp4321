package retrieval

import (
	"comp4321/stopword"
	"strings"
	"github.com/surgebase/porter2"
	"comp4321/models"
	"comp4321/database"
	"log"
	"math"
	"sort"
	"regexp"
)

func preprocessText(words []string) (rv []string) {
	regex := regexp.MustCompile("[^a-zA-Z0-9 ]")
	for _, word := range words {
		cleaned := regex.ReplaceAllString(word, " ")
		cleaned = strings.TrimSpace(word)
		cleaned = strings.ToLower(cleaned)
		cleaned = porter2.Stem(cleaned)
		if !stopword.IsStopWord(cleaned) {
			rv = append(rv, cleaned)
		}
	}
	return
}

type SEngine struct {
	viewer *database.Viewer
}

func NewSearchEngine(filename string) *SEngine {
	se := SEngine{}
	var err error
	se.viewer, err = database.LoadViewer(filename)
	if err != nil {
		log.Fatal("Index file not found:", filename)
	}

	return &se;
}

func (e *SEngine) getDocumentViewModels(docIds []uint64, scores map[uint64]float64) []*models.DocumentView {
	rv := make([]*models.DocumentView, len(docIds))
	for i, id := range docIds {
		doc := e.viewer.GetDocument(id)
		if doc != nil {
			docView := models.NewDocumentView(doc)
			if scores == nil {
				docView.Score = 1;
			} else {
				docView.Score = scores[id];
			}
			parents := e.viewer.GetParentLinks(id)
			upper := int(math.Min(float64(len(parents)), 5.0))
			docView.Parents = parents[0:upper]
			rv[i] = docView
		}
	}
	return rv
}

func (e *SEngine) RetrieveBoolean(query string) []*models.DocumentView {
	e.viewer, _ = database.LoadViewer("index.db")
	querySplit := strings.Split(query, " ")
	preprocessed := preprocessText(querySplit)
	docIds := booleanFilter(preprocessed, e.viewer)

	return e.getDocumentViewModels(docIds, nil)
}

func (e *SEngine) RetrievePhrase(query string) []*models.DocumentView {
	e.viewer, _ = database.LoadViewer("index.db")
	querySplit := strings.Split(query, " ")
	preprocessed := preprocessText(querySplit)
	docIds := searchPhrase(preprocessed, e.viewer)
	scores, ids := getDocumentScores(preprocessed, e.viewer, docIds)

	sort.Slice(docIds, func(i, j int) bool {
		return scores[ids[i]] > scores[ids[j]]
	})

	upper := int(math.Min(50.0, float64(len(ids))))
	ids = ids[0:upper]

	return e.getDocumentViewModels(ids, scores)
}

func (e *SEngine) RetrieveVSpace(query string) []*models.DocumentView {
	e.viewer, _ = database.LoadViewer("index.db")
	querySplit := strings.Split(query, " ")
	preprocessed := preprocessText(querySplit)
	scores, docIds := vspaceRetrieval(preprocessed, e.viewer)

	return e.getDocumentViewModels(docIds, scores)
}

func (e *SEngine) Close() {
	e.viewer.Close()
}
