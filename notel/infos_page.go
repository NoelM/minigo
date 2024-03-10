package main

import (
	"github.com/NoelM/minigo"
)

func NewPageInfo(mntl *minigo.Minitel) *minigo.Page {
	infoPage := minigo.NewPage("notel:info", mntl, nil)

	var items []Depeche
	var pageStartItem = []int{0}
	var pageId int

	infoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		items = LoadFeed(France24FeedURL)

		printInfoHeader(mntl)
		pageStartItem = append(pageStartItem, printDepeche(mntl, items[pageStartItem[pageId]:], 4))
		printInfoHelpers(mntl)

		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	infoPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		// We verify that the postId on the next page exists
		if pageStartItem[pageId+1] >= len(items)-1 {
			return nil, minigo.NoOp
		}
		mntl.CleanScreen()

		pageId += 1
		deltaId := printDepeche(mntl, items[pageStartItem[pageId]:], 1)

		if pageId+1 == len(pageStartItem) {
			// If the pageId+1 has never been loaded, increase the pageStart array
			pageStartItem = append(pageStartItem, pageStartItem[pageId]+deltaId)
		} else {
			// The page already been loaded, but the data may have changed.
			// So we update the pageStart
			pageStartItem[pageId+1] = pageStartItem[pageId] + deltaId
		}
		printInfoHelpers(mntl)

		return nil, minigo.NoOp
	})

	infoPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageId == 0 {
			return nil, minigo.NoOp
		}
		mntl.CleanScreen()

		pageId -= 1
		if pageId == 0 {
			printInfoHeader(mntl)
			printDepeche(mntl, items[pageStartItem[pageId]:], 4)
		} else {
			printDepeche(mntl, items[pageStartItem[pageId]:], 1)
		}
		printInfoHelpers(mntl)

		return nil, minigo.NoOp
	})

	return infoPage
}

func printDepeche(mntl *minigo.Minitel, depeches []Depeche, startLine int) int {
	const maxLine = 22

	line := startLine
	mntl.MoveAt(line, 0)
	var trunc bool

	var dId int
	var d Depeche
	for dId, d = range depeches {
		// Display Date
		if line+1 > maxLine {
			break
		}

		mntl.ModeG0()
		mntl.Attributes(minigo.FondBleu, minigo.CaractereBlanc)
		mntl.Print(" " + d.Date.Format("02/01/2006 15:04"))
		mntl.SendCAN()
		mntl.Return(1)
		line += 1

		// Display Title
		title := minigo.WrapperLargeurNormale(d.Title)
		if line+len(title) > maxLine {
			trunc = true
			break
		}

		for _, l := range title {
			mntl.Attributes(minigo.FondCyan, minigo.CaractereNoir)
			mntl.Print(" " + l)
			mntl.SendCAN()
			mntl.Return(1)

			line += 1
		}

		// Display Content
		content := minigo.WrapperLargeurNormale(d.Content)

		mntl.Attributes(minigo.CaractereBlanc, minigo.FondNormal)
		for _, l := range content {
			mntl.Print(l)
			line += 1
			if line > maxLine {
				trunc = true
				break
			}
			mntl.Return(1)
		}

		mntl.Return(1)
		line += 1

		// Yes, a message is at least 3 lines
		// - Date Line
		// - Title Line
		// - Content Line
		// - Blank Line
		//
		// So at maxLine-4, it would automatically break
		if line >= maxLine-4 {
			break
		}
	}

	if !trunc {
		// The last post has not been truncated, we'll load the next one
		dId += 1
	}
	return dId

}

func printInfoHeader(mntl *minigo.Minitel) {
	mntl.MoveAt(2, 1)
	mntl.Attributes(minigo.DoubleHauteur)
	mntl.Print("Dépèches France 24")
	mntl.Attributes(minigo.GrandeurNormale)
}

func printInfoHelpers(mntl *minigo.Minitel) {
	mntl.HelperLeftAt(24, "Menu", "SOMMAIRE")
	mntl.HelperRightAt(24, "Naviguez", "SUITE/RETOUR")
}
