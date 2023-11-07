package main

import (
	"github.com/mmcdole/gofeed"
	"time"
)

const France24FeedURL = "https://www.france24.com/fr/rss"

type Depeche struct {
	Title    string
	Category string
	Date     *time.Time
	Content  string
}

func LoadFeed(url string) (dep []Depeche) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)

	for _, i := range feed.Items {
		dep = append(dep, Depeche{
			Title:    i.Title,
			Category: i.Categories[0],
			Date:     i.PublishedParsed,
			Content:  i.Description,
		})
	}

	return
}
