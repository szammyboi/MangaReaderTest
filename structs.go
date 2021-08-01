package main

type AllManga struct {
	Manga []Series `json:"manga"`
}

type Series struct {
	Title         string    `json:"title"`
	VanityTitle   string    `json:"vanityTitle"`
	LatestChapter string    `json:"latestChapter"`
	Thumbnail     string    `json:"thumbnail"`
	Chapters      []Chapter `json:"chapters"`
}

type Chapter struct {
	Chapter  string `json:"chapter"`
	MangaID  string `json:"mangaID"`
	DataURL  string `json:"dataURL"`
	ImageURL string `json:"imageURL"`
}

type ImageLinks struct {
	Links []string `json:"links"`
}
