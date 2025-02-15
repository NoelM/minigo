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
	msgStopRow  = 20
	hLineRow    = 21
	inputRow    = 22
	helpersRow  = 24
)

const (
	NoLimit  = -1
	MaxLimit = 256
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
	c.mntl.MoveAt(hLineRow, 0)
	c.mntl.CleanLine()

	c.mntl.Return(helpersRow - hLineRow)
	c.mntl.CleanLine()
}

func (c *ChatLayout) printFooter() {
	c.mntl.MoveAt(hLineRow, 0)
	c.mntl.HLine(40, minigo.HCenter)
	// It already went to the next line!
	c.mntl.Return(helpersRow - 1 - hLineRow)

	c.mntl.Helper("Aide →", "GUIDE", minigo.FondBleu, minigo.CaractereBlanc)
	c.mntl.HelperRight("→", "ENVOI", minigo.FondVert, minigo.CaractereNoir)
}

func (c *ChatLayout) printHeader() {
	cntd := c.cntd.Load()

	mode := "EDITION"
	if c.navMode {
		mode = "NAVIGATION"
	}

	var status string
	if cntd < 2 {
		status = fmt.Sprintf(" [Mode %s] Connecté: %d", mode, cntd)
	} else {
		status = fmt.Sprintf(" [Mode %s] Connectés: %d", mode, cntd)
	}
	c.mntl.PrintStatusWithAttributes(fmt.Sprintf("%-35s", status), minigo.FondMagenta, minigo.CaractereNoir)
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
		c.mntl.Return(1)
		c.cache.AppendBottom(Date, msgId, 0)

	} else if dir == Up {
		c.mntl.ReturnUp(1)
		c.cache.AppendTop(Date, msgId, 0)
	}

	c.mntl.Attributes(minigo.CaractereBleu)
	c.mntl.PrintCenter(dateString)
	c.mntl.Attributes(minigo.CaractereBlanc)

	return 1
}

func (c *ChatLayout) printMessageFromLine(msgId, lineId, limit int, dir RouleauDir) int {
	lines, vdt := FormatMessage(c.messages[msgId], dir, c.mntl.SupportCSI())

	if limit < 0 || limit > lines {
		limit = lines
	}

	start := lineId
	// we already reach the end of the msg
	if start > lines-1 {
		return 0
	}

	for i := start; i < limit; i += 1 {
		c.mntl.Send(vdt[i])

		if dir == Up {
			c.cache.AppendTop(Message, msgId, i)
		} else if dir == Down {
			c.cache.AppendBottom(Message, msgId, i)
		}
	}

	return limit
}

func (c *ChatLayout) printMessage(msgId, limit int, dir RouleauDir) int {
	return c.printMessageFromLine(msgId, 0, limit, dir)
}

func (c *ChatLayout) PrintPreviousMessage() {
	if !c.navMode {
		c.mntl.CursorOff()
		c.navMode = true
		c.printHeader()
	}

	firstRow := c.cache.FirstRow()

	var msgId int
	if firstRow.kind == Date {
		msgId = firstRow.msgId - 1
		if msgId <= 0 {
			return
		}

		c.mntl.MoveAt(1, 0)
		c.printMessage(msgId, NoLimit, Up)
		c.printDate(msgId, NoLimit, Up)

	} else {
		c.mntl.MoveAt(1, 0)
		msgId = firstRow.msgId

		c.printMessageFromLine(msgId, firstRow.lineId+1, NoLimit, Up)
		c.printDate(msgId, NoLimit, Up)

		msgId -= 1
		if msgId <= 0 {
			return
		}

		c.printMessage(msgId, NoLimit, Up)
		c.printDate(msgId, NoLimit, Up)
	}
}

func (c *ChatLayout) PrintNextMessage() {
	rowId, lastRow := c.cache.LastRow()
	if lastRow.msgId == len(c.messages)-1 {
		return
	}

	if !c.navMode {
		c.mntl.CursorOff()
		c.navMode = true
		c.printHeader()
	}

	msgId := lastRow.msgId
	c.mntl.MoveAt(rowId+1, 0)

	if lastRow.kind == Date {
		c.printMessage(msgId, NoLimit, Down)

	} else {
		c.printMessageFromLine(msgId, lastRow.lineId+1, NoLimit, Down)

		msgId += 1
		if msgId == len(c.messages)-1 {
			return
		}

		c.printDate(msgId, NoLimit, Down)
		c.printMessage(msgId, NoLimit, Down)
	}
}

func (c *ChatLayout) Init() {
	// Load the last messages from DB
	if !c.getLastMessages() && !c.navMode {
		// No message, we quit!
		return
	}

	c.navMode = false
	c.cache.Init()

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
	if c.navMode {
		c.mntl.CleanScreen()
		c.Init()
		return
	}

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
