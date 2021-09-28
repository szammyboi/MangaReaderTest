package main

type AllManga struct {
	Manga []Series `json:"manga"`
}

type Series struct {
	Title         string    `json:"title"`
	VanityTitle   string    `json:"vanityTitle"`
	LatestChapter string    `json:"latestChapter"`
	Thumbnail     string    `json:"thumbnail"`
	Saved         bool      `json:"saved"`
	Chapters      []Chapter `json:"chapters"`
}

type Chapter struct {
	Chapter  string `json:"chapter"`
	MangaID  string `json:"mangaID"`
	DataURL  string `json:"dataURL"`
	ImageURL string `json:"imageURL"`
	Saved    bool   `json:"saved"`
}

type ImageLinks struct {
	Links []string `json:"links"`
}

type ChapterInfo struct {
	Info []ChapterInfoNode `json:"info"`
}

type ChapterInfoNode struct {
	URL string `json:"url"`
	Key string `json:"key"`
}
