package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var updatedToday bool
var sess *session.Session
var uploader *s3manager.Uploader
var bucket string

func mainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func update(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var writeData []byte

	client := &http.Client{}
	mangaMinReq, requestErr := http.NewRequest("GET", "https://d2j9ticyfssj97.cloudfront.net/Series/mangamin.json", nil)
	if requestErr != nil {
		log.Fatal(requestErr)
	}

	mangaMin, responseErr := client.Do(mangaMinReq)
	if responseErr != nil {
		log.Fatal(responseErr)
	}

	if mangaMin.StatusCode == http.StatusOK {
		writeData, _ = ioutil.ReadAll(mangaMin.Body)
	}

	// fix timing here bruv
	if !updatedToday {
		fmt.Println("Updated DB")
		writeData = updateDatabase(writeData)
		updatedToday = true
	} else if updatedToday && time.Now().Hour() < 12 {
		updatedToday = false
	}

	fmt.Println(time.Since(start))

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(writeData)
}

func seriesJSON(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	series := vars["series"]

	selectedSeries := fetchSeriesJSON(series, false)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(toJSON(selectedSeries))
}

func getChapter(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	series := vars["series"]
	chapter := vars["chapter"]

	fetchChapter(series, chapter)
	baseURL := fmt.Sprintf("https://d2j9ticyfssj97.cloudfront.net/Series/%s/%s/%s", series, chapter, chapter+".json")
	client := &http.Client{}
	fmt.Println(baseURL)
	mangaInfoReq, requestErr := http.NewRequest("GET", baseURL, nil)
	if requestErr != nil {
		log.Fatal(requestErr)
	}

	mangaInfoResp, responseErr := client.Do(mangaInfoReq)
	if responseErr != nil {
		log.Fatal(responseErr)
	}

	if mangaInfoResp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(mangaInfoResp.Body)
		fmt.Printf("Fetching %s Chapter %s... %s\n", series, chapter, fmt.Sprint(time.Since(start)))
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(bodyBytes)
	}
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeFile(w, r, "./frontend/"+file)
}

func main() {
	updatedToday = false
	sess, _ = session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	uploader = s3manager.NewUploader(sess)

	bucket = os.Getenv("S3_BUCKET")
	webRouter := mux.NewRouter().StrictSlash(true)
	webRouter.HandleFunc("/", mainPage)
	webRouter.HandleFunc("/update", update)

	webRouter.HandleFunc("/json/{series}", seriesJSON)
	webRouter.HandleFunc("/getChapter/{series}/{chapter}", getChapter)
	webRouter.HandleFunc("/getPage/{series}/{chapter}/{page}", getPage)

	webRouter.HandleFunc("/reader/{series}", reader)
	webRouter.HandleFunc("/assets/{file}", assets)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "1000"
	}
	log.Fatal(http.ListenAndServe(":"+PORT, webRouter))
}
