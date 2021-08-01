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

func main() {
	webRouter := mux.NewRouter().StrictSlash(true)
	webRouter.HandleFunc("/", mainPage)

	PORT := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(":"+PORT, webRouter))
}
