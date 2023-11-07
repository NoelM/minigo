package main

import (
	"github.com/NoelM/minigo"
	"strings"
	"unicode/utf8"
)

func NewPageInfo(mntl *minigo.Minitel, pseudo string) *minigo.Page {
	infoPage := minigo.NewPage("notel:info", mntl, nil)

	var depeches []Depeche
	var pageStart = []int{0}
	var pageId int

	infoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		depeches = LoadFeed(France24FeedURL)
		pageStart = append(pageStart, printDepeche(mntl, depeches[pageStart[pageId]:]))

		mntl.WriteHelperLeft(24, "Menu NOTEL", "SOMMAIRE")
		mntl.WriteHelperLeft(24, "Naviguez", "SUITE/RETOUR")

		return minigo.NoOp
	})

	infoPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	infoPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	infoPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {

		// We verify that the postId on the next page exists
		if pageStart[pageId+1] >= len(depeches)-1 {
			return nil, minigo.NoOp
		}

		pageId += 1
		deltaId := printDepeche(mntl, depeches[pageStart[pageId]:])

		if pageId+1 == len(pageStart) {
			// If the pageId+1 has never been loaded, increase the pageStart array
			pageStart = append(pageStart, pageStart[pageId]+deltaId)
		} else {
			// The page already been loaded, but the data may have changed.
			// So we update the pageStart
			pageStart[pageId+1] = pageStart[pageId] + deltaId
		}

		return nil, minigo.NoOp
	})

	infoPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageId == 0 {
			return nil, minigo.NoOp
		}

		pageId -= 1
		printDepeche(mntl, depeches[pageStart[pageId]:])

		return nil, minigo.NoOp
	})

	return infoPage
}

func printDepeche(mntl *minigo.Minitel, posts []Depeche) int {
	line := 1
	var trunc bool

	var pId int
	var p Depeche
	for pId, p = range posts {
		mntl.WriteAttributes(minigo.InversionFond)
		mntl.WriteStringLeft(line, p.Pseudo)
		mntl.WriteStringRight(line, p.Date.Format("02/01/06 15:04"))
		mntl.WriteAttributes(minigo.FondNormal)

		line += 1
		if line == 22 {
			trunc = true
			break
		}

		length := 0
		words := []string{}
		for _, s := range strings.Split(p.Content, " ") {
			if length+utf8.RuneCountInString(s)+1 >= 40 {
				mntl.WriteStringLeft(line, strings.Join(words, " "))

				length = 0
				words = []string{}

				line += 1
				if line == 22 {
					trunc = true
					break
				}
			}

			length += utf8.RuneCountInString(s) + 1 // the size of the space
			words = append(words, s)
		}

		// Yes, a message is at least 2 lines
		// - Status Line
		// - Message Line
		//
		// So at line=20, it would automatically break
		if line >= 20 {
			break
		}
	}

	if !trunc {
		// The last post has not been truncated, we'll load the next one
		pId += 1
	}
	return pId

}
