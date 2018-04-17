package main

import (
	"comp4321/models"
	"comp4321/retrieval"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
)

var homeTemplate = template.Must(template.ParseFiles("views/home.html"))
var resultTemplate = template.Must(template.ParseFiles("views/results.html"))

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, nil)
	fmt.Println(r.URL.String())
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	viewModel := models.ResultView{}
	queries := r.URL.Query().Get("keywords")
	page := r.URL.Query().Get("page")
	currentPage, err := strconv.Atoi(page)

	if err != nil {
		currentPage = 1
	}

	se := retrieval.NewSearchEngine("index.db")
	defer se.Close()
	viewModel.Query = queries
	viewModel.Results = se.RetrievePhrase(queries)

	pageNum := math.Ceil(float64(len(viewModel.Results)) / 10.0)
	if currentPage > int(pageNum) {
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
	maxindex := int(math.Min(float64(currentPage*10), float64(len(viewModel.Results))))
	viewModel.Results = viewModel.Results[(currentPage-1)*10 : maxindex]
	viewModel.CurrentPage = currentPage
	viewModel.PageNum = int(pageNum)

	resultTemplate.Execute(w, viewModel)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	// Load indexer

	// File servers
	staticServer := http.FileServer(http.Dir("static"))
	viewServer := http.FileServer(http.Dir("views"))

	// Handle requests for routes
	http.HandleFunc("/", helloWorldHandler)
	http.Handle("/views/", http.StripPrefix("/views/", viewServer))
	http.Handle("/static/", http.StripPrefix("/static/", staticServer))
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)

	fmt.Println(http.ListenAndServe(":8080", nil))
}
