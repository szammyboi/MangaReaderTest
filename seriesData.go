/*

__/\\\\\\\\\\\\\\\_______/\\\\\_______/\\\\\\\\\\\\__________/\\\\\______
 _\///////\\\/////______/\\\///\\\____\/\\\////////\\\______/\\\///\\\____
  _______\/\\\_________/\\\/__\///\\\__\/\\\______\//\\\___/\\\/__\///\\\__
   _______\/\\\________/\\\______\//\\\_\/\\\_______\/\\\__/\\\______\//\\\_
    _______\/\\\_______\/\\\_______\/\\\_\/\\\_______\/\\\_\/\\\_______\/\\\_
     _______\/\\\_______\//\\\______/\\\__\/\\\_______\/\\\_\//\\\______/\\\__
      _______\/\\\________\///\\\__/\\\____\/\\\_______/\\\___\///\\\__/\\\____
       _______\/\\\__________\///\\\\\/_____\/\\\\\\\\\\\\/______\///\\\\\/_____
        _______\///_____________\/////_______\////////////__________\/////_______


        * POSSIBLY REVERT TO THE OLD WAY FOR INCREASED SPEED OR FIND A WAY TO VALIDATE IF THEY NEED TO BE UPDATED *
        * IMPLEMENT A TRUE UPDATING SYSTEM THAT DELETES THE DETAIL FILES FOR MANGAS THAT WERE PREVIOUSLY UPDATED AND THEN CALL FETCHCHAPTERS ON THEM *
        * MAYBE MAKE IT SO THAT ALL THE CHAPTERS FROM THE FULL FILE ARE PUT INTO THE SAME ORDER AS THE MIN FILE SO THAT THEY ARE IN LATEST RELEASE ORDER *
        * PASS THE FILE AS A POINTER TO THE UPDATE CHAPTER FUNCTION FOR EFFICIENCY *
        * IMPLEMENT NEW SERIES ADDITION ALGORITHM *
        * ADD A CHAPTER PAGE LENGTH SECTION TO THE SERVER QUERY TO ELIMINATE THE RISKY METHOD OF FETCHING PAGES *
        * REWRITE READER COMPLETELY *


*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gocolly/colly"
)

func fetchSeriesData() {
	mangaList := make([]Series, 0)

	baseSearch := colly.NewCollector(
		colly.Async(true),
	)

	baseSearch.OnHTML(".o_sortable", func(e *colly.HTMLElement) {
		title := e.ChildText(".type-md--lg")
		link := e.ChildAttr(".o_chapters-link", "href")
		icon := e.ChildAttr("img", "data-original")
		vanity := link[21:]

		latest := strings.Split(e.ChildText("span"), "\n")[0]
		latest = strings.ReplaceAll(latest, "Latest: Chapter ", "")

		newManga := Series{
			Title:         title,
			VanityTitle:   vanity,
			LatestChapter: latest,
			Thumbnail:     icon,
			Chapters:      make([]Chapter, 0),
		}

		mangaList = append(mangaList, newManga)
	})

	baseSearch.Visit("https://www.viz.com/shonenjump")

	baseSearch.Wait()

	allManga := AllManga{
		Manga: mangaList,
	}

	saveToJSON(allManga, "manga_min.json")
}

func returnSeriesData() []Series {
	mangaList := make([]Series, 0)

	baseSearch := colly.NewCollector(
		colly.Async(true),
	)

	baseSearch.OnHTML(".o_sortable", func(e *colly.HTMLElement) {
		title := e.ChildText(".type-md--lg")
		link := e.ChildAttr(".o_chapters-link", "href")
		icon := e.ChildAttr("img", "data-original")
		vanity := link[21:]

		latest := strings.Split(e.ChildText("span"), "\n")[0]
		latest = strings.ReplaceAll(latest, "Latest: Chapter ", "")

		newManga := Series{
			Title:         title,
			VanityTitle:   vanity,
			LatestChapter: latest,
			Thumbnail:     icon,
			Chapters:      make([]Chapter, 0),
		}

		mangaList = append(mangaList, newManga)
	})

	baseSearch.Visit("https://www.viz.com/shonenjump")

	baseSearch.Wait()

	return mangaList
}

func updateDatabase(rawJSON []byte) []byte {
	freshScan := returnSeriesData()
	var dbManga AllManga
	json.Unmarshal(rawJSON, &dbManga)
	database := dbManga.Manga

	// vanity titles then call update chapter on that
	// or implement latest chapter attr here probalbyl the most effi
	updateQueue := make([]Series, 0)
	for _, manga := range freshScan {
		for _, savedManga := range database {
			if manga.VanityTitle == savedManga.VanityTitle {
				if manga.LatestChapter != savedManga.LatestChapter {
					updateQueue = append(updateQueue, manga)
				}
				break
			}
		}
		// new series
	}

	allManga := AllManga{
		Manga: freshScan,
	}

	if len(updateQueue) > 0 {
		updateChapters(updateQueue)
	}

	json := toJSON(allManga)
	key := "Series/mangamin.json"
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

	return json
	//saveToJSON(allManga, "manga_min.json")
}

func updateChapters(updateQueue []Series) {
	var dbChapters = loadMangaFromAWS("Series/mangafull.json")
	var wg sync.WaitGroup
	var found bool
	for _, newSeries := range updateQueue {
		found = false
		for index, oldSeries := range dbChapters {
			if newSeries.VanityTitle == oldSeries.VanityTitle {
				wg.Add(1)
				dbChapters[index].LatestChapter = newSeries.LatestChapter
				go updateSeries(index, newSeries.VanityTitle, &dbChapters, &wg)
			}
		}
		if !found {
			wg.Add(1)
			dbChapters = append(dbChapters, newSeries)
			go updateSeries(len(dbChapters)-1, newSeries.VanityTitle, &dbChapters, &wg)
		}
	}
	wg.Wait()
	json := toJSON(AllManga{Manga: dbChapters})
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
	fmt.Println("ALL UPDATED")
	fmt.Println(result)

}

func updateSeries(selectedSeries int, vanityTitle string, db *[]Series, wg *sync.WaitGroup) {

	chapters := make([]Chapter, 0)
	foundChapters := make([]string, 0)

	chapterSearch := colly.NewCollector(
		colly.Async(false),
	)

	chapterSearch.OnHTML("[data-target-url]", func(e *colly.HTMLElement) {
		target := e.Attr("data-target-url")
		target = target[strings.Index(target, "/"):]
		target = strings.ReplaceAll(target, "');", "")

		values := strings.Split(target, "/")

		chapter := values[2]
		chapter = chapter[strings.Index(chapter, "chapter-")+len("chapter-"):]
		//chapter = strings.ReplaceAll(chapter, "-", ".")

		for _, foundChapter := range foundChapters {
			if foundChapter == chapter {
				return
			}
		}

		values[4] = strings.ReplaceAll(values[4], "?action=read", "")
		mangaID := values[4]

		newChapter := Chapter{
			Chapter:  chapter,
			MangaID:  mangaID,
			DataURL:  "https://www.viz.com" + target,
			ImageURL: "https://www.viz.com/manga/get_manga_url?device_id=3&manga_id=" + mangaID + "&page=",
			Saved:    false,
		}

		chapters = append(chapters, newChapter)
		foundChapters = append(foundChapters, chapter)
	})

	chapterSearch.Visit("https://viz.com/shonenjump/chapters/" + vanityTitle)
	chapterSearch.Wait()

	(*db)[selectedSeries].Chapters = chapters
	wg.Done()
}
