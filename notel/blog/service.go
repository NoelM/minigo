package blog

import (
	"encoding/json"
	"os"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/logs"
)

type Article struct {
	Title   string `json:"title"`
	Date    string `json:"date"`
	Content string `json:"content"`
}

func ServiceBlog(m *minigo.Minitel, blogDbPath string) int {
	articles := []Article{}

	if data, err := os.ReadFile(blogDbPath); err != nil {
		logs.ErrorLog("unable to open Blog DB: %s\n", err)
		return minigo.SommaireOp

	} else {
		if err := json.Unmarshal(data, &articles); err != nil {
			logs.ErrorLog("unable to unmarshal Blog DB: %s\n", err)
			return minigo.SommaireOp
		}
	}

	articleId := 0
	if len(articles) == 0 {
		logs.ErrorLog("the blog DB is empty, leaving\n")
		return minigo.SommaireOp
	}

DISPLAY:
	_, op := NewArticlePage(m, articles[articleId]).Run()
	switch op {
	case minigo.SuiteOp:
		articleId += 1
		if articleId >= len(articles) {
			articleId = len(articles) - 1
		}
		goto DISPLAY

	case minigo.RetourOp:
		articleId -= 1
		if articleId < 0 {
			articleId = 0
		}
		goto DISPLAY

	default:
		return op
	}
}
