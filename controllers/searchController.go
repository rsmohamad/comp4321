package controllers

import (
	"net/http"
	"comp4321/models"
	"comp4321/retrieval"
	"html/template"
	"time"
	"log"
	"fmt"
)

var resultTemplate *template.Template

func searchHandler(w http.ResponseWriter, r *http.Request) {
	viewModel := models.ResultView{}
	queries := r.URL.Query().Get("keywords")

	startSearch := time.Now()
	se := retrieval.NewSearchEngine("index.db")
	viewModel.Query = queries
	viewModel.Results = se.RetrievePhrase(queries)
	viewModel.TotalResults = len(viewModel.Results)
	se.Close()
	elapsed := time.Since(startSearch)

	log.Println(fmt.Sprintf("Query: \"%s\" %s", queries, elapsed))
	resultTemplate.ExecuteTemplate(w, "resultView", viewModel)
}

func LoadSearch(){
	resultTemplate, _ = template.ParseFiles("views/resultView.html", "views/documentView.html")
	http.HandleFunc("/search/", searchHandler)
}