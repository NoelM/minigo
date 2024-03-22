package infos

import (
	"fmt"

	"github.com/NoelM/minigo"
)

func NewPageInfo(mntl *minigo.Minitel, service map[string]string) *minigo.Page {
	infoPage := minigo.NewPage("notel:info", mntl, nil)

	var items []Depeche
	var itemId int

	infoPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.Reset()
		mntl.RouleauOn()

		items = LoadFeed(service["url"])

		printInfoHeader(mntl, service["name"])
		printInfoHelpers(mntl)

		printDepeche(mntl, items[itemId])
		itemId += 1

		return minigo.NoOp
	})

	infoPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		mntl.RouleauOff()
		return nil, minigo.SommaireOp
	})

	infoPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if itemId >= len(items) {
			itemId = len(items)
			return nil, minigo.NoOp
		}

		printDepeche(mntl, items[itemId])
		itemId += 1

		return nil, minigo.NoOp
	})

	infoPage.SetGuideFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		printInfoHelpers(mntl)
		return nil, minigo.NoOp
	})

	return infoPage
}

func printDepeche(mntl *minigo.Minitel, d Depeche) {
	mntl.Return(2)

	mntl.ModeG0()
	mntl.Attributes(minigo.FondBlanc, minigo.CaractereRouge)
	mntl.Print(" " + d.Date.Format("02/01/2006 15:04"))
	mntl.SendCAN()
	mntl.Return(1)

	// Display Title
	title := minigo.WrapperLargeurNormale(d.Title)

	for _, l := range title {
		mntl.Attributes(minigo.FondRouge, minigo.CaractereBlanc)
		mntl.Print(" " + l)
		mntl.SendCAN()
		mntl.Return(1)
	}

	// Display Content
	content := minigo.WrapperGenerique(d.Content, 38)

	mntl.Attributes(minigo.CaractereBlanc, minigo.FondNormal)
	for _, l := range content {
		mntl.ReturnCol(1, 1)
		mntl.Print(l)
	}
}

func printInfoHeader(mntl *minigo.Minitel, name string) {
	mntl.MoveAt(2, 0)
	mntl.Attributes(minigo.DoubleHauteur)
	mntl.PrintCenter(fmt.Sprintf("-- %s --", name))
	mntl.Attributes(minigo.GrandeurNormale)
}

func printInfoHelpers(mntl *minigo.Minitel) {
	mntl.Return(1)
	mntl.HLine(40, minigo.HCenter)

	mntl.Right(1)
	mntl.PrintAttributes("SUITE", minigo.InversionFond)
	mntl.Right(2)
	mntl.Print("→ Info suivante")

	mntl.ReturnCol(1, 1)
	mntl.PrintAttributes("SOMMAIRE", minigo.InversionFond)
	mntl.Right(2)
	mntl.Print("→ Choix des sources")

	mntl.ReturnCol(1, 1)
	mntl.PrintAttributes("GUIDE", minigo.InversionFond)
	mntl.Right(2)
	mntl.Print("→ Réafficher cette aide")

	mntl.Return(1)
	mntl.HLine(40, minigo.HCenter)
}
