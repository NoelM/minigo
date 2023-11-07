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

		printHeader(mntl)
		pageStartItem = append(pageStartItem, printDepeche(mntl, items[pageStartItem[pageId]:], 4))

		mntl.WriteHelperLeft(24, "Menu NOTEL", "SOMMAIRE")
		mntl.WriteHelperRight(24, "Naviguez", "SUITE/RETOUR")

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
		if pageStartItem[pageId+1] >= len(items)-1 {
			return nil, minigo.NoOp
		}

		if pageId == 0 {
			mntl.CleanNRowsFrom(1, 1, 23)
		}

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

		return nil, minigo.NoOp
	})

	infoPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageId == 0 {
			return nil, minigo.NoOp
		}

		pageId -= 1
		if pageId == 0 {
			printHeader(mntl)
			printDepeche(mntl, items[pageStartItem[pageId]:], 4)
		} else {
			printDepeche(mntl, items[pageStartItem[pageId]:], 1)
		}

		return nil, minigo.NoOp
	})

	return infoPage
}

func printDepeche(mntl *minigo.Minitel, depeches []Depeche, startLine int) int {
	const maxLine = 23

	line := startLine
	var trunc bool

	// Clean lines from startLine to 23
	mntl.CleanNRowsFrom(startLine, 1, maxLine-startLine+1)

	var dId int
	var d Depeche
	for dId, d = range depeches {
		// Display Date
		if line+1 > maxLine {
			break
		}
		mntl.WriteStringRight(line, d.Date.Format("02/01/2006 15:04"))
		line += 1

		// Display Title
		title := minigo.WrapperGenerique(d.Title, 38)
		if line+len(title) > maxLine {
			trunc = true
			break
		}

		for _, l := range title {
			mntl.WriteAttributes(minigo.DebutLignage)
			mntl.WriteStringLeft(line, " "+l)
			line += 1
		}

		// Display Content
		content := minigo.WrapperLargeurNormale(d.Content)

		for _, l := range content {
			mntl.WriteStringLeft(line, l)
			line += 1
			if line > maxLine {
				trunc = true
				break
			}
		}

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

func printHeader(mntl *minigo.Minitel) {
	mntl.WriteAttributes(minigo.DoubleHauteur)
	mntl.WriteStringLeft(2, "Infos de FRANCE24")
	mntl.WriteAttributes(minigo.GrandeurNormale)
}
