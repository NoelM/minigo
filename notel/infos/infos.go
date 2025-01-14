package infos

import (
	"time"

	"github.com/mmcdole/gofeed"
)

const France24Rss = "https://www.france24.com/fr/rss"
const FranceInfoRss = "https://www.francetvinfo.fr/titres.rss"
const LeMondeRss = "https://www.lemonde.fr/rss/une.xml"
const LeMondeLiveRss = "https://www.lemonde.fr/rss/en_continu.xml"
const BBCRss = "http://feeds.bbci.co.uk/news/rss.xml"
const LiberationRss = "https://www.liberation.fr/arc/outboundfeeds/rss-all/?outputType=xml"
const TheVergeRss = "https://www.theverge.com/rss/index.xml"
const LeFigaroRss = "https://www.lefigaro.fr/rss/figaro_flash-actu.xml"

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
