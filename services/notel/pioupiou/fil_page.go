package pioupiou

import (
	"strings"
	"unicode/utf8"

	"github.com/NoelM/minigo"
)

func NewPageFil(mntl *minigo.Minitel, pseudo string) *minigo.Page {
	filPage := minigo.NewPage("pioupiou:message", mntl, nil)

	var posts []*PPPost
	var pageStart = []int{0}
	var pageId int

	filPage.SetInitFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, initData map[string]string) int {
		mntl.CleanScreen()

		var err error
		if posts, err = PiouPiou.GetFeed(pseudo); err != nil {
			errorLog.Printf("failed to get feed: %s\n", err.Error())
		}

		pageStart = append(pageStart, printPosts(mntl, posts[pageStart[pageId]:]))

		mntl.WriteHelperLeft(23, "Naviguez", "SUITE/RETOUR")
		mntl.WriteHelperRight(23, "Rafraichir", "REPET")

		mntl.WriteHelperLeft(24, "Menu", "SOMMAIRE")
		inputs.AppendInput("command", minigo.NewInput(mntl, 24, 29, 4, 1, true))
		mntl.WriteHelperRight(24, ".... +", "ENVOI")

		inputs.InitAll()
		return minigo.NoOp
	})

	filPage.SetEnvoiFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		cmdString := inputs.ValueActive()
		if len(cmdString) == 0 {
			return nil, minigo.NoOp
		}

		cmdId := ParseCommandString(string(cmdString))
		if cmdId < 0 {
			PrintErrorMessage("Commande inconnue, utilisez GUIDE")
		}
		return nil, cmdId
	})

	filPage.SetCorrectionFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		inputs.CorrectionActive()
		return nil, minigo.NoOp
	})

	filPage.SetCharFunc(func(mntl *minigo.Minitel, inputs *minigo.Form, key int32) {
		inputs.AppendKeyActive(key)
	})

	filPage.SetSommaireFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		return nil, minigo.SommaireOp
	})

	filPage.SetSuiteFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageStart[pageId+1] >= len(posts)-1 {
			tmp, err := PiouPiou.GetFeed(pseudo)
			if err != nil {
				return nil, minigo.NoOp
			}
			if len(tmp) == 0 {
				return nil, minigo.NoOp
			}

			posts = append(posts, tmp...)
		}

		pageId += 1
		deltaId := printPosts(mntl, posts[pageStart[pageId]:])

		if pageId+1 == len(pageStart) {
			pageStart = append(pageStart, pageStart[pageId]+deltaId)
		} else {
			pageStart[pageId+1] = pageStart[pageId] + deltaId
		}

		return nil, minigo.NoOp
	})

	filPage.SetRetourFunc(func(mntl *minigo.Minitel, inputs *minigo.Form) (map[string]string, int) {
		if pageId-1 < 0 {
			return nil, minigo.NoOp
		}
		printPosts(mntl, posts[pageStart[pageId-1]:])
		pageId -= 1

		return nil, minigo.NoOp
	})

	return filPage
}

func printPosts(mntl *minigo.Minitel, posts []*PPPost) int {
	line := 1
	var trunc bool

	var pId int
	var p *PPPost
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
		pId += 1
	}
	return pId

}
