package superchat

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

type RouleauDir uint

const (
	Up   RouleauDir = 0
	Down RouleauDir = 1
)

const (
	msgStartRow = 1
	msgStopRow  = 19
	hLineRow    = 21
	inputRow    = 22
	helpersRow  = 24
)

const (
	Blank = -2
	Date  = -1
)

const NoLimit = -1

type ChatLayout struct {
	mntl *minigo.Minitel

	msgDB    *databases.MessageDatabase
	messages []databases.Message
	maxId    int

	nick string

	navMode bool
	cache   *Cache

	cntd *atomic.Int32
}

func NewChatLayout(mntl *minigo.Minitel, msgDB *databases.MessageDatabase, cntd *atomic.Int32, nick string) *ChatLayout {
	return &ChatLayout{
		mntl:  mntl,
		msgDB: msgDB,
		maxId: -1,
		nick:  nick,
		cache: NewCache(),
		cntd:  cntd,
	}
}

func (c *ChatLayout) cleanFooter() {
	c.mntl.MoveAt(hLineRow, 0)
	c.mntl.CleanLine()

	//c.mntl.Return(helpersRow - 1 - hLineRow)
	//c.mntl.CleanLine()

	c.mntl.Return(1)
	c.mntl.CleanLine()
}

func (c *ChatLayout) printFooter() {
	c.mntl.MoveAt(hLineRow, 0)
	c.mntl.HLine(40, minigo.HCenter)
	// It already went to the next line!
	c.mntl.Return(helpersRow - 1 - hLineRow)

	c.mntl.Helper("Aide", "GUIDE", minigo.FondBleu, minigo.CaractereBlanc)
	c.mntl.HelperRight("→", "ENVOI", minigo.FondVert, minigo.CaractereNoir)
}

func (c *ChatLayout) printHeader() {
	cntd := c.cntd.Load()

	if cntd < 2 {
		c.mntl.PrintStatus(fmt.Sprintf("→ Connecté: %d", cntd))
	} else {
		c.mntl.PrintStatus(fmt.Sprintf("→ Connectés: %d", cntd))
	}
}

func (c *ChatLayout) getLastMessages() bool {
	if last := c.msgDB.GetMessages(c.nick); len(last) == 0 {
		return false

	} else {
		c.messages = append(c.messages, last...)
		return true
	}
}

func (c *ChatLayout) printDate(msgId, limit int, dir RouleauDir) int {
	if limit >= 0 && limit < 1 {
		return 0
	}

	var lastDate time.Time
	if msgId-1 > 0 {
		lastDate = c.messages[msgId-1].Time
	}

	dateString := GetDateString(lastDate, c.messages[msgId].Time)
	if dateString == "" {
		return 0
	}

	if dir == Down {
		//c.mntl.Return(1)
		//c.cache.Bottom(Blank)

		c.mntl.Return(1)
		c.cache.Bottom(Date)

	} else if dir == Up {
		c.mntl.ReturnUp(1)
		c.cache.Top(Date)
	}

	c.mntl.Attributes(minigo.CaractereBleu)
	c.mntl.PrintCenter(dateString)
	c.mntl.Attributes(minigo.CaractereBlanc)

	//if dir == Up {
	//	c.mntl.ReturnUp(1)
	//	c.cache.Top(Blank)
	//}

	return 1
}

func (c *ChatLayout) printMessage(msgId, limit int, dir RouleauDir) int {
	lines, vdt := FormatMessage(c.messages[msgId], dir, c.mntl.SupportCSI())

	if limit < 0 || limit > lines {
		limit = lines
	}

	for i := 0; i < limit; i += 1 {
		c.mntl.Send(vdt[i])

		if dir == Up {
			c.cache.Top(msgId)
		} else if dir == Down {
			c.cache.Bottom(msgId)
		}
	}

	return limit
}

func (c *ChatLayout) Init() {
	// Load the last messages from DB
	if !c.getLastMessages() {
		// No message, we quit!
		return
	}

	// No cursor and go to the origin
	c.mntl.CursorOff()
	c.mntl.MoveAt(1, 0)

	// We'll use the rouleau mode from the TOP
	// Until one reaches the `rowMsgZoneEnd`
	curLine := 0

	// Limit means the number of avail lines until rowMsgZoneEnd
	limit := 0

	// One start from the last message recvd.
	for msgId := len(c.messages) - 1; msgId >= 0; msgId -= 1 {
		limit = msgStopRow - curLine
		curLine += c.printMessage(msgId, limit, Up)

		limit = msgStopRow - curLine
		curLine += c.printDate(msgId, limit, Up)

		if curLine >= msgStopRow {
			break
		}
	}
	c.maxId = len(c.messages) - 1

	c.printFooter()
	c.printHeader()
}

func (c *ChatLayout) Update() {
	if !c.getLastMessages() {
		return
	}

	// Clean screen before the update
	c.mntl.CursorOff()
	c.cleanFooter()

	// Go to the last line of the MSG Zone
	c.mntl.MoveAt(msgStopRow, 1)

	// We print on the DOWN direction all the new messages, no limits here!
	curLine := msgStopRow
	for msgId := c.maxId + 1; msgId < len(c.messages); msgId += 1 {
		curLine += c.printDate(msgId, NoLimit, Down)
		curLine += c.printMessage(msgId, NoLimit, Down)
	}
	c.maxId = len(c.messages) - 1

	// If the new line is below 24 (the last on screen)
	// Move the cursor there, otherwise, the rouleau mode
	// will not push blank lines to `endMsgZone`
	if curLine < 24 {
		c.mntl.MoveAt(24, 0)
	}

	// Now push curLine to rowMsgZoneEng
	// this is a > not a >= because
	// with an equal at curLine == rowMsgZoneEnd
	// if will return another time
	for ; curLine > msgStopRow; curLine -= 1 {
		c.mntl.Return(1)
	}

	c.printHeader()
	c.printFooter()
}
