package main

import (
	"fmt"
	"net/http"
	"html/template"
	"comp4321/database"
	"comp4321/models"
)

var homeTemplate = template.Must(template.ParseFiles("views/home.html"))
var resultTemplate = template.Must(template.ParseFiles("views/results.html"))
var viewer *database.Viewer

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, nil)
	fmt.Println(r.URL.String())
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	keys, _ := r.URL.Query()["keywords"]
	fmt.Println("Received query for " + keys[0])

	viewModel := models.ResultView{}
	viewModel.Query = keys[0]

	viewer.ForEachDocument(func(p *models.Document, i int) {
		viewModel.Results = append(viewModel.Results, p)
	})

	resultTemplate.Execute(w, viewModel)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	// Load indexer
	viewer, _ = database.LoadViewer("index.db")

	// File servers
	staticServer := http.FileServer(http.Dir("static"))
	viewServer := http.FileServer(http.Dir("views"))

	// Handle requests for routes
	http.HandleFunc("/", helloWorldHandler)
	http.Handle("/views/", http.StripPrefix("/views/", viewServer))
	http.Handle("/static/", http.StripPrefix("/static/", staticServer))
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)

	fmt.Println(http.ListenAndServe(":80", nil))
}
