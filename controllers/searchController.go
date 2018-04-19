package controllers

import (
	"net/http"
	"comp4321/models"
	"comp4321/retrieval"
	"html/template"
	"time"
	"log"
	"fmt"
	"comp4321/database"
)

var resultTemplate *template.Template

func nestedHandler(w http.ResponseWriter, r *http.Request) {
	viewModel := models.ResultView{}
	haystack := r.URL.Query().Get("haystack")
	needle := r.URL.Query().Get("needle")

	startSearch := time.Now()
	se := retrieval.NewSearchEngine("index.db")
	viewModel.Query = fmt.Sprintf("<%s> INSIDE <%s>", needle, haystack)
	viewModel.Results = se.RetrieveNested(haystack, needle)
	viewModel.TotalResults = len(viewModel.Results)
	se.Close()
	elapsed := time.Since(startSearch)

	log.Println(fmt.Sprintf("[%s] [%s] [%s]", r.RemoteAddr, viewModel.Query, elapsed))
	resultTemplate.ExecuteTemplate(w, "resultView", viewModel)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	viewModel := models.ResultView{}
	queries := r.URL.Query().Get("keywords")

	userId := database.GetCookieInstance().GetCookieId(r)
	database.GetCookieInstance().SetCookieResponse(userId, w)
	database.GetCookieInstance().AddQuery(userId, queries)

	startSearch := time.Now()
	se := retrieval.NewSearchEngine("index.db")
	viewModel.Query = queries
	viewModel.Results = se.RetrievePhrase(queries)
	viewModel.TotalResults = len(viewModel.Results)
	se.Close()
	elapsed := time.Since(startSearch)

	log.Println(fmt.Sprintf("[%s] [%s] [%s]", r.RemoteAddr, queries, elapsed))
	resultTemplate.ExecuteTemplate(w, "resultView", viewModel)
}

func LoadSearch() {
	resultTemplate, _ = template.ParseFiles("views/resultView.html", "views/documentView.html")
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/search/nested/", nestedHandler)
}
