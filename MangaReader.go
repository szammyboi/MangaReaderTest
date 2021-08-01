package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var updatedToday bool

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

func getChapter(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	series := vars["series"]
	chapter := vars["chapter"]

	folder, _ := ioutil.ReadDir(fmt.Sprintf("./Series/%s/Chapters/%s", series, chapter))
	if len(folder) == 0 {
		fetchChapter(series, chapter)
		folder, _ = ioutil.ReadDir(fmt.Sprintf("./Series/%s/Chapters/%s", series, chapter))
	}

	alllinks := ImageLinks{}
	baseURL := fmt.Sprintf("http://chemistry-tutor.com/getPage/%s/%s/", series, chapter)
	for i := 0; i < len(folder); i++ {
		alllinks.Links = append(alllinks.Links, fmt.Sprintf(baseURL+"%d.jpg", i))
	}

	fmt.Printf("Fetching %s Chapter %s... %s\n", series, chapter, fmt.Sprint(time.Since(start)))
	w.Header().Set("Content-Type", "application/json")
	w.Write(toJSON(alllinks))
}

func reader(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./frontend/index.html")
}

func getPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	series := vars["series"]
	chapter := vars["chapter"]
	pageImage := vars["page"]

	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeFile(w, r, fmt.Sprintf("./Series/%s/Chapters/%s/%s", series, chapter, pageImage))
}

func assets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]
	http.ServeFile(w, r, "./frontend/"+file)
}

func main() {
	updatedToday = false
	webRouter := mux.NewRouter().StrictSlash(true)
	webRouter.HandleFunc("/", mainPage)
	webRouter.HandleFunc("/update", update)

	webRouter.HandleFunc("/json/{series}", seriesJSON)
	webRouter.HandleFunc("/getChapter/{series}/{chapter}", getChapter)
	webRouter.HandleFunc("/getPage/{series}/{chapter}/{page}", getPage)

	webRouter.HandleFunc("/reader/{series}", reader)
	webRouter.HandleFunc("/assets/{file}", assets)

	PORT := os.Getenv("PORT")

	log.Fatal(http.ListenAndServe(":"+PORT, webRouter))
}
