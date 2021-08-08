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
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

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

func updateDatabase() {
	freshScan := returnSeriesData()
	jsonFile, _ := os.Open("manga_min.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var dbManga AllManga
	json.Unmarshal(byteValue, &dbManga)
	database := dbManga.Manga

	// vanity titles then call update chapter on that
	updateQueue := make([]string, 0)
	for _, manga := range freshScan {
		for _, savedManga := range database {
			if manga.VanityTitle == savedManga.VanityTitle {
				if manga.LatestChapter != savedManga.LatestChapter {
					updateQueue = append(updateQueue, manga.VanityTitle)
				}
				break
			}
		}
		// new series
	}

	if len(updateQueue) > 0 {
		for _, s := range updateQueue {
			updateChapter(s)
		}
	}

	allManga := AllManga{
		Manga: freshScan,
	}

	saveToJSON(allManga, "manga_min.json")
}

func updateChapter(vanityTitle string) {

	jsonFile, err := os.Open("manga_full.json")

	if err != nil {
		log.Println(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var savedManga AllManga
	var selectedSeries int
	json.Unmarshal(byteValue, &savedManga)
	database := savedManga.Manga

	for seriesIndex, currentSeries := range database {
		if currentSeries.VanityTitle == vanityTitle {
			selectedSeries = seriesIndex
			break
		}
	}

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

	database[selectedSeries].Chapters = chapters

	if len(chapters) > 0 {
		if chapters[0].Chapter == "1" {
			database[selectedSeries].LatestChapter = chapters[len(chapters)-1].Chapter
		} else {
			database[selectedSeries].LatestChapter = chapters[0].Chapter
		}
	}

	updatedManga := AllManga{
		Manga: database,
	}

	os.Mkdir("Series/"+vanityTitle, 0644)
	os.Mkdir("Series/"+vanityTitle+"/Chapters", 0644)
	saveToJSON(updatedManga, "manga_full.json")

}
