package main

import (
	"comp4321/controllers"
	"net/http"
	"log"
)

func main() {
	controllers.LoadHome()
	controllers.LoadSearch()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
