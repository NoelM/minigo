package minichat

import (
	"fmt"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

type RouleauDir uint

const (
	Up   RouleauDir = 0
	Down RouleauDir = 1
)

const (
	rowMsgZoneStart = 1
	rowMsgZoneEnd   = 19
	rowHLine        = 20
	rowInput        = 21
	rowHelpers      = 24
)

const (
	Blank = -2
	Date  = -1
)

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
	c.mntl.MoveCursorAt(rowHLine, 1)
	c.mntl.CleanLine()

	c.mntl.MoveCursorAt(rowHelpers-1, 1)
	c.mntl.CleanLine()

	c.mntl.MoveCursorAt(24, 1)
	c.mntl.CleanLine()
}

func (c *ChatLayout) printFooter() {
	c.mntl.HLine(rowHLine, 1, 40, minigo.HCenter)

	c.mntl.HLine(rowHelpers-1, 1, 40, minigo.HCenter)
	c.mntl.WriteHelperLeft(rowHelpers, "Màj. écran", "REPET.")
	c.mntl.WriteHelperRight(rowHelpers, "Message +", "ENVOI")
}

func (c *ChatLayout) printHeader() {
	cntd := c.cntd.Load()

	if cntd < 2 {
		c.mntl.WriteStatusLine(fmt.Sprintf("> Connecté: %d", cntd))
	} else {
		c.mntl.WriteStatusLine(fmt.Sprintf("> Connectés: %d", cntd))
	}
}

func (c *ChatLayout) getLastMessages() bool {
	if last := c.msgDB.GetMessages(c.nick); last == nil {
		return false

	} else {
		c.messages = append(c.messages, last...)
		return true
	}
}

func (c *ChatLayout) printDate(msgId, limit int, dir RouleauDir) int {
	if limit < 2 {
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
		c.mntl.Return(1)
		c.cache.Bottom(Blank)

		c.mntl.Return(1)
		c.cache.Bottom(Date)

	} else if dir == Up {
		c.mntl.ReturnUp(1)
		c.cache.Top(Date)
	}

	c.mntl.WriteAttributes(minigo.CaractereBleu)

	length := utf8.RuneCountInString(dateString)
	c.mntl.MoveCursorRight((minigo.ColonnesSimple - length) / 2)
	c.mntl.WriteString(dateString)

	c.mntl.WriteAttributes(minigo.CaractereBlanc)

	if dir == Up {
		c.mntl.ReturnUp(1)
		c.cache.Top(Blank)
	}

	return 2
}

func (c *ChatLayout) printMessage(msgId, limit int, dir RouleauDir) int {
	lines, vdt := FormatMessage(c.messages[msgId], dir)

	if limit < 1 || limit > lines {
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
	c.mntl.CursorOff()

	// Load the last messages from DB
	if !c.getLastMessages() {
		return
	}

	c.mntl.MoveCursorAt(1, 1)

	curLine := 0
	limit := rowMsgZoneEnd - curLine
	for msgId := len(c.messages) - 1; msgId >= 0; msgId -= 1 {
		curLine += c.printMessage(msgId, limit, Up)

		limit = rowMsgZoneEnd - curLine
		curLine += c.printDate(msgId, limit, Up)

		if curLine == rowMsgZoneEnd {
			break
		}
	}
	c.maxId = len(c.messages) - 1

	c.printFooter()
	c.printHeader()
}
