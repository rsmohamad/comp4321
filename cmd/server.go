package main

import (
	"fmt"
	"net/http"
	"html/template"
	"comp4321/database"
	"comp4321/models"
	"strconv"
	"math"
)

var homeTemplate = template.Must(template.ParseFiles("views/home.html"))
var resultTemplate = template.Must(template.ParseFiles("views/results.html"))
var viewer *database.Viewer

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, nil)
	fmt.Println(r.URL.String())
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	viewModel := models.ResultView{}
	keys := r.URL.Query().Get("keywords")
	page := r.URL.Query().Get("page")
	currentPage, err := strconv.Atoi(page)
	if err != nil {
		currentPage = 1
	}
	viewModel.Query = keys
	viewer.ForEachDocument(func(p *models.Document, i int) {
		viewModel.Results = append(viewModel.Results, p)
	})

	pageNum := math.Ceil(float64(len(viewModel.Results)) / 10.0)
	if currentPage > int(pageNum){
		currentPage = 1
	}

	// 10 pages pagination window
	viewModel.Pages = make([]int, 0)
	min := math.Max(float64(currentPage-5), 1)
	max := math.Min(min+9, pageNum)
	for i := int(min); i <= int(max); i++ {
		viewModel.Pages = append(viewModel.Pages, i)
	}

	viewModel.TotalResults = len(viewModel.Results)
	viewModel.Results = viewModel.Results[(currentPage-1)*10: currentPage*10]
	viewModel.CurrentPage = currentPage;
	viewModel.PageNum = int(pageNum)

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
