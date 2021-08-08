package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
