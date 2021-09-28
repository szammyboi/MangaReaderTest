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
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gocolly/colly"
)

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
	database := loadMangaFromAWS("Series/mangamin.json")

	updateQueue := make([]Series, 0)
	for freshPos, manga := range freshScan {
		for savedPos, savedManga := range database {
			if manga.VanityTitle == savedManga.VanityTitle {
				freshScan[freshPos].Saved = database[savedPos].Saved
				if manga.LatestChapter != savedManga.LatestChapter && savedManga.Saved {
					fmt.Println(manga.Title, "Old:", savedManga.LatestChapter, "New: ", manga.LatestChapter)
					updateQueue = append(updateQueue, manga)
				}
				break
			}
		}
	}

	allManga := AllManga{
		Manga: freshScan,
	}

	var wg sync.WaitGroup
	fmt.Println("UPDATE QUEUE:")
	for _, chap := range updateQueue {
		wg.Add(1)
		updateSeriesJSON(chap.VanityTitle, true, &wg)
	}
	wg.Wait()

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
}

func fetchSeriesJSON(series string, update bool) *Series {
	var saved Series
	min := loadMangaFromAWS("Series/mangamin.json")
	isSaved := false

	if min == nil {
		fmt.Println("AWS FAILED")
		return &Series{}
	}
	for pos, manga := range min {
		if manga.VanityTitle == series {
			if manga.Saved {
				saved = loadSeriesFromAWS("Series/" + series + "/series.json")
				isSaved = true
				if !update {
					return &saved
				}
			}
			min[pos].Saved = true
			break
		}
	}

	selectedSeries := findManga(&min, series)

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

	chapterSearch.Visit("https://viz.com/shonenjump/chapters/" + series)
	chapterSearch.Wait()

	selectedSeries.Chapters = chapters

	if isSaved {
		for i, newChapter := range selectedSeries.Chapters {
			for _, savedChapter := range saved.Chapters {
				if newChapter.Chapter == savedChapter.Chapter {
					selectedSeries.Chapters[i].Saved = savedChapter.Saved
					break
				}
			}
		}
	}

	json := toJSON(selectedSeries)
	key := "Series/" + selectedSeries.VanityTitle + "/series.json"
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

	json2 := toJSON(AllManga{Manga: min})
	key2 := "Series/mangamin.json"
	test3 := bytes.NewReader(json2)
	upParams = &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key2,
		Body:   test3,
	}

	result2, uploaderr2 := uploader.Upload(upParams)
	if uploaderr2 != nil {
		log.Fatal(uploaderr)
	}
	fmt.Println(result2)
	return selectedSeries
}

func updateSeriesJSON(series string, update bool, wg *sync.WaitGroup) {
	fetchSeriesJSON(series, true)
	wg.Done()
}
