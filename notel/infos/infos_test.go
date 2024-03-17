package infos

import (
	"testing"
)

func TestLoadFeed(t *testing.T) {
	feed := LoadFeed(France24Rss)

	if len(feed) == 0 {
		t.Fatal("empty feed")
	}

	f := feed[0]
	t.Log(f.Title)
	t.Log(f.Date)
	t.Log(f.Category)
	t.Log(f.Content)
}
