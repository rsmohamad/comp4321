package main

import (
	"github.com/rsmohamad/comp4321/controllers"
	"net/http"
	"log"
)

func main() {
	controllers.LoadHome()
	controllers.LoadSearch()
	controllers.LoadHistory()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
