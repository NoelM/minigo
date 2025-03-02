package blog

import (
	"encoding/json"
	"os"
	"strconv"

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

LIST:
	choice, op := ListPage(m, articles).Run()
	if op != minigo.EnvoiOp {
		return op
	}

	articleId, err := strconv.Atoi(choice["article"])
	if err != nil {
		logs.ErrorLog("unable to convert article id to int: %s\n", err)
		goto LIST
	}
	articleId -= 1

	if articleId < 0 || articleId >= len(articles) {
		logs.ErrorLog("invalid article id: %d\n", articleId)
		goto LIST
	}

DISPLAY:
	_, op = NewArticlePage(m, articles[articleId]).Run()
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
