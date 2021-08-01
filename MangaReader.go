package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var updatedToday bool

func hiddenPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "hidden.html")
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func update(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if !updatedToday && time.Now().Hour() >= 12 {
		fmt.Println("Updated DB")
		updateDatabase()
		updatedToday = true
	} else if updatedToday {
		updatedToday = false
	}
	fmt.Println(time.Since(start))

	http.ServeFile(w, r, "manga_min.json")
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
	updatedToday = false
	webRouter := mux.NewRouter().StrictSlash(true)
	webRouter.HandleFunc("/", hiddenPage)
	webRouter.HandleFunc("/browse", mainPage)
	webRouter.HandleFunc("/update", update)
	webRouter.HandleFunc("/json/{series}", seriesJSON)
	PORT := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(":"+PORT, webRouter))
}
