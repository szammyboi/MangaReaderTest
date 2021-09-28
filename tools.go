package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func toJSON(t interface{}) []byte {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	err := encoder.Encode(t)
	if err != nil {
		fmt.Println("FAILED JSON ENCODING")
		return nil
	}

	return buffer.Bytes()
}

func printJSON(t interface{}) string {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	err := encoder.Encode(t)
	if err != nil {
		fmt.Println("FAILED JSON ENCODING")
		return ""
	}

	return string(buffer.Bytes())
}

func saveToJSON(t interface{}, filename string) {
	_ = ioutil.WriteFile(filename, toJSON(t), 0644)
}

func loadManga(file string) []Series {
	f, _ := os.Open(file)
	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)

	var p AllManga
	json.Unmarshal(byteValue, &p)

	return p.Manga
}

func loadMangaFromAWS(filename string) []Series {

	client := &http.Client{}
	mangareq, requestErr := http.NewRequest("GET", "https://d2j9ticyfssj97.cloudfront.net/"+filename, nil)
	if requestErr != nil {
		log.Fatal(requestErr)
	}

	manga, responseErr := client.Do(mangareq)
	if responseErr != nil {
		log.Fatal(responseErr)
	}

	if manga.StatusCode == http.StatusOK {
		mangadata, readerr := ioutil.ReadAll(manga.Body)
		if readerr != nil {
			log.Fatal(readerr)
		}

		var p AllManga
		json.Unmarshal(mangadata, &p)
		return p.Manga
	} else {
		log.Fatal("ERROR!")
		log.Fatal("STATUS CODE:", manga.StatusCode)
	}
	log.Fatal("Something went wrong")
	return make([]Series, 0)
}

func loadSeriesFromAWS(filename string) Series {

	client := &http.Client{}
	mangareq, requestErr := http.NewRequest("GET", "https://d2j9ticyfssj97.cloudfront.net/"+filename, nil)
	if requestErr != nil {
		log.Fatal(requestErr)
	}

	manga, responseErr := client.Do(mangareq)
	if responseErr != nil {
		log.Fatal(responseErr)
	}

	if manga.StatusCode == http.StatusOK {
		mangadata, readerr := ioutil.ReadAll(manga.Body)
		if readerr != nil {
			log.Fatal(readerr)
		}
		var p Series
		json.Unmarshal(mangadata, &p)

		return p
	} else {
		log.Fatal("ERROR!")
		log.Fatal("STATUS CODE:", manga.StatusCode)
	}
	return Series{}
}

func findManga(database *[]Series, tag string) *Series {
	for _, savedSeries := range *database {
		if savedSeries.VanityTitle == tag {
			return &savedSeries
		}
	}
	return nil
}

func findChapter(chapters *[]Chapter, tag string) *Chapter {
	for _, savedChapter := range *chapters {
		if savedChapter.Chapter == tag {
			return &savedChapter
		}
	}
	return nil
}

func findChapterAndPosition(chapters *[]Chapter, tag string) (*Chapter, int) {
	for i, savedChapter := range *chapters {
		if savedChapter.Chapter == tag {
			return &savedChapter, i
		}
	}
	return nil, 0
}

func exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		/*exists*/
		return true
	} else { /*not exists or some other error*/
		return false
	}
}
