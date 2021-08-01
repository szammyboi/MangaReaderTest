package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func mainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func seriesJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	series := vars["series"]

	database := loadManga("manga_full.json")
	selectedSeries := findManga(&database, series)

	w.Header().Set("Content-Type", "application/json")
	w.Write(toJSON(selectedSeries))
}

func main() {
	webRouter := mux.NewRouter().StrictSlash(true)
	webRouter.HandleFunc("/", mainPage)
	webRouter.HandleFunc("/json/{series}", seriesJSON)
	PORT := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(":"+PORT, webRouter))
}
