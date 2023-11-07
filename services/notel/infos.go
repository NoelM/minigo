package main

import (
	"time"

	"github.com/mmcdole/gofeed"
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
		category := ""
		if len(i.Categories) > 0 {
			category = i.Categories[0]
		}

		dep = append(dep, Depeche{
			Title:    i.Title,
			Category: category,
			Date:     i.PublishedParsed,
			Content:  i.Description,
		})
	}

	return
}
