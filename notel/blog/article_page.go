package blog

import "github.com/NoelM/minigo"

func NewArticlePage(mntl *minigo.Minitel, article Article) *minigo.Page {
	page := minigo.NewPage("blog:article", mntl, nil)

	wrappedTextId := 0
	wrappedText := minigo.Wrapper(article.Content, 38)

	page.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()
		mntl.CursorOff()

		mntl.MoveAt(3, 1)
		mntl.Attributes(minigo.DoubleGrandeur)
		mntl.Print(article.Title)
		mntl.Attributes(minigo.GrandeurNormale)

		mntl.Return(1)
		mntl.HLine(40, minigo.HCenter)

		for i := 5; i < 24; i++ {
			mntl.ReturnCol(1, 1)

			if wrappedTextId == len(wrappedText) {
				break
			}

			mntl.Print(wrappedText[wrappedTextId])
			wrappedTextId += 1
		}

		return minigo.NoOp
	})

	page.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if wrappedTextId == len(wrappedText) {
			return nil, minigo.SuiteOp
		}

		mntl.CleanScreenFrom(5, 0)
		mntl.MoveAt(4, 0)

		for i := 5; i < 24; i++ {
			mntl.ReturnCol(1, 1)

			if wrappedTextId == len(wrappedText) {
				break
			}

			mntl.Print(wrappedText[wrappedTextId])
			wrappedTextId += 1
		}

		return nil, minigo.NoOp
	})

	page.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if wrappedTextId == 0 {
			return nil, minigo.RetourOp
		}

		wrappedTextId -= 2 * (24 - 5)
		if wrappedTextId < 0 {
			wrappedTextId = 0
		}

		mntl.CleanScreenFrom(5, 0)
		mntl.MoveAt(4, 0)

		for i := 5; i < 24; i++ {
			mntl.ReturnCol(1, 1)

			if wrappedTextId == len(wrappedText) {
				break
			}

			mntl.Print(wrappedText[wrappedTextId])
			wrappedTextId += 1
		}

		return nil, minigo.NoOp
	})

	page.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	return page
}
