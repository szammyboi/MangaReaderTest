package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

var cookie string

type imageData struct {
	URL  string
	Page int
}

func fetchChapter(specifiedSeries string, specifiedChapter string) {
	exif.RegisterParsers(mknote.All...)
	database := loadMangaFromAWS("Series/" + specifiedSeries + ".json")
	selectedSeries := findManga(&database, specifiedSeries)
	selectedChapter, pos := findChapterAndPosition(&selectedSeries.Chapters, specifiedChapter)
	client := &http.Client{}

	pageCount := fetchPageCount(client, selectedChapter)
	//os.Mkdir("Series/"+selectedSeries.VanityTitle, 0644)
	//os.Mkdir("Series/"+selectedSeries.VanityTitle+"/Chapters", 0644)
	if selectedChapter.Saved {
		return
	}
	cookie = "announcement_init=true; _fbp=fb.1.1609807369275.633303156; _pin_unauth=dWlkPU5UVTJOamhsTVRVdE5UbGtZeTAwTmpsbExUaGlaalV0T0RjNVpEQTFZek5rTVdOaQ; __gads=ID=63aa8b18be10224e:T=1609807397:S=ALNI_MZtlWpj0hbu9IXuT9Ws1Ndc3GKnjw; __stripe_mid=5d7a5d46-2136-4711-87c2-cfa3a6cd5359c59d43; curtain_seen=true; chapter-series-694-follow-modal=2021-05-01; property-1671-follow-modal=2021-05-01; calendar_filter=manga-books; calendar_view=product-table; chapter-series-724-follow-modal=2021-05-02; chapter-series-540-follow-modal=2021-05-02; _session_id=96a89d1ced081206349d92042ddbe13a; chapter-series-448-follow-modal=2021-06-21; _derived_epik=dj0yJnU9T25meVAzNVpoQy1pSHhMQjZ4TUxFZ28yQ0IxbmF6QW4mbj1DX1dERHBTV2Z1QzVCTjMzYjlXV0pBJm09NyZ0PUFBQUFBR0RXUDRVJnJtPWUmcnQ9QUFBQUFHQkpTa1k; pixlee_analytics_cookie_legacy=%7B%22CURRENT_PIXLEE_USER_ID%22%3A%221208a0fd-733a-91b8-9702-b07329a38f31%22%7D; pixlee_analytics_cookie=%7B%22CURRENT_PIXLEE_USER_ID%22%3A%221208a0fd-733a-91b8-9702-b07329a38f31%22%7D; property-1875-follow-modal=2021-07-03; _gcl_au=1.1.1669550766.1625364968; chapter-series-739-follow-modal=2021-07-04; property-2267-follow-modal=2021-07-05; chapter-series-781-follow-modal=2021-07-05; chapter-series-5-follow-modal=2021-07-06; chapter-series-716-follow-modal=2021-07-08; chapter-series-249-follow-modal=2021-07-08; chapter-series-553-follow-modal=2021-07-08; chapter-series-699-follow-modal=2021-07-08; chapter-series-520-follow-modal=2021-07-09; _gid=GA1.2.1980397515.1626130283; chapter-series-722-follow-modal=2021-07-13; iter_id=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb21wYW55X2lkIjoiNWU1NjllNWFmMjMzMmMwMDAxMTE2NDRjIiwidXNlcl9pZCI6IjYwMGRkMjI5Y2FhM2E3MDAwMTQxMDRkZiIsImlhdCI6MTYyNjE0OTg4NH0.kQ6Ar55U_iAgXzHxpjOZa7VCHOSwvevEy6byB4n7gSI; chapter-series-790-follow-modal=2021-07-13; user_visits=1; user_visits_url=https%3A%2F%2Fwww.viz.com%2Fshonenjump%2Fone-piece-chapter-1010%2Fchapter%2F22351%3Faction%3Dread; _gat=1; _gat_UA-136373-5=1; remember_token=fYX9eUg9VxCraf4ux9yFdeh3AdNRmBnJ; _ga_41C9NY052Q=GS1.1.1626194402.187.1.1626194428.0; _ga=GA1.1.1087606157.1609807369"

	fetchImages(client, selectedSeries, selectedChapter, pageCount)

	selectedSeries.Chapters[pos].Saved = true
	json := toJSON(AllManga{Manga: database})
	key := "Series/mangafull.json"
	test2 := bytes.NewReader(json)
	upParams := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   test2,
	}

	result, uploaderr := uploader.Upload(upParams)
	if uploaderr != nil {
		log.Fatal(uploaderr)
	}
	fmt.Println(result)
}

func fetchImages(client *http.Client, selectedSeries *Series, selectedChapter *Chapter, pageCount int) {
	linkChan := make(chan imageData, pageCount+1)
	keyStorage := make([]ChapterInfoNode, pageCount+1)

	var wg sync.WaitGroup
	for i := 0; i <= pageCount; i++ {
		wg.Add(1)
		go fetchImageLink(client, selectedChapter, i, linkChan)
		go fetchPage(client, linkChan, &wg, selectedSeries, selectedChapter, &keyStorage)
	}
	wg.Wait()
	close(linkChan)

	keys := ChapterInfo{
		Info: keyStorage,
	}

	key := fmt.Sprintf("Series/%s/%s/%s", selectedSeries.VanityTitle, selectedChapter.Chapter, selectedChapter.Chapter+".json")
	test2 := bytes.NewReader(toJSON(keys))
	upParams := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   test2,
	}

	result, uploaderr := uploader.Upload(upParams)
	if uploaderr != nil {
		log.Fatal(uploaderr)
	}
	fmt.Println(result)
}

func fetchImageLink(client *http.Client, selectedChapter *Chapter, page int, buffer chan imageData) {
	imageLinkRequest, requestErr := http.NewRequest("GET", selectedChapter.ImageURL+strconv.Itoa(page), nil)
	if requestErr != nil {
		fmt.Println("imageLinkRequest Error")
		log.Fatal(requestErr)
	}

	imageLinkRequest.Header.Add("referer", selectedChapter.DataURL)
	imageLinkRequest.Header.Add("cookie", cookie)

	imageLinkResponse, responseErr := client.Do(imageLinkRequest)
	if responseErr != nil {
		fmt.Println("imageLinkResponse Error")
		log.Fatal(responseErr)
	}

	if imageLinkResponse.StatusCode == http.StatusOK {
		responseBytes, _ := ioutil.ReadAll(imageLinkResponse.Body)
		buffer <- imageData{URL: string(responseBytes), Page: page}
	}
}

func fetchPageCount(client *http.Client, currentChapter *Chapter) int {
	pageRequest, _ := http.NewRequest("GET", currentChapter.DataURL, nil)
	pageRequest.Header.Add("cookie", "_session_id=96a89d1ced081206349d92042ddbe13a;")
	pageResponse, _ := client.Do(pageRequest)

	pageBytes, _ := ioutil.ReadAll(pageResponse.Body)
	pageStr := string(pageBytes)
	pageStr = pageStr[strings.Index(pageStr, "var pages")+len("var pages"):]
	pageStr = pageStr[:strings.Index(pageStr, ";")]
	pageStr = strings.ReplaceAll(pageStr, " ", "")
	pageStr = strings.ReplaceAll(pageStr, "=", "")

	count, _ := strconv.Atoi(pageStr)
	return count
}

func fetchPage(client *http.Client, linkChan chan imageData, wg *sync.WaitGroup, selectedSeries *Series, selectedChapter *Chapter, keyStorage *[]ChapterInfoNode) {
	req := <-linkChan
	url := req.URL
	page := req.Page

	imageRequest, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		fmt.Println("imageRequest Error")
		log.Fatal(requestErr)
	}

	imageResponse, responseErr := client.Do(imageRequest)
	if responseErr != nil {
		fmt.Println("imageResponse Error")
		log.Fatal(responseErr)
	}

	if imageResponse.StatusCode == http.StatusOK {
		imageBytes, imageErr := ioutil.ReadAll(imageResponse.Body)
		if imageErr != nil {
			log.Fatal(imageErr)
		}

		key := fmt.Sprintf("Series/%s/%s/%s", selectedSeries.VanityTitle, selectedChapter.Chapter, strconv.Itoa(page)+".jpg")
		keyReader := bytes.NewReader(imageBytes)
		uploadReader := bytes.NewReader(imageBytes)

		tags, _ := exif.Decode(keyReader)
		decodekey, _ := tags.Get(exif.ImageUniqueID)
		decodestring, _ := decodekey.StringVal()
		imageURL := fmt.Sprintf("https://d2j9ticyfssj97.cloudfront.net/Series/%s/%s/%s", selectedSeries.VanityTitle, selectedChapter.Chapter, strconv.Itoa(page)+".jpg")
		(*keyStorage)[page] = ChapterInfoNode{URL: imageURL, Key: decodestring}

		upParams := &s3manager.UploadInput{
			Bucket: &bucket,
			Key:    &key,
			Body:   uploadReader,
		}

		result, uploaderr := uploader.Upload(upParams)
		if uploaderr != nil {
			log.Fatal(uploaderr)
		}
		fmt.Println(result)
		//os.Mkdir("Series/"+selectedSeries.VanityTitle+"/Chapters/"+selectedChapter.Chapter, 0644)
		//ioutil.WriteFile("Series/"+selectedSeries.VanityTitle+"/Chapters/"+selectedChapter.Chapter+"/"+page+".jpg", imageBytes, 0644)
	}
	wg.Done()
}
