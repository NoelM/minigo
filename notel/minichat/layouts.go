package minichat

import (
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

type ChatLayout struct {
	mntl *minigo.Minitel

	msgDB    *databases.MessageDatabase
	messages []databases.Message
	maxId    int

	nick string

	rowEndMsgZone int
	rowHLine      int
	rowInput      int

	navMode bool
	cache   *Cache
}

func NewChatLayout(mntl *minigo.Minitel, msgDB *databases.MessageDatabase, nick string, inputLine, inputHeight int) *ChatLayout {
	return &ChatLayout{
		mntl:          mntl,
		msgDB:         msgDB,
		maxId:         -1,
		nick:          nick,
		rowEndMsgZone: 20,
		rowHLine:      21,
		rowInput:      22,
		cache:         NewCache(),
	}
}

func (c *ChatLayout) cleanHelpers() {
	c.mntl.MoveCursorAt(c.rowHLine, 1)
	c.mntl.CleanLine()

	c.mntl.MoveCursorAt(24, 1)
	c.mntl.CleanLine()
}

func (c *ChatLayout) printHelpers() {
	c.mntl.HLine(c.rowHLine, 1, 40, minigo.HCenter)

	c.mntl.WriteHelperLeft(24, "Màj. écran", "REPET.")
	c.mntl.WriteHelperRight(24, "Message +", "ENVOI")
}

func (c *ChatLayout) getLastMessages() bool {
	if last := c.msgDB.GetMessages(c.nick); last == nil {
		return false

	} else {
		c.messages = append(c.messages, last...)
		return true
	}
}

func (c *ChatLayout) printDate(msgId int) int {
	var lastDate time.Time
	if msgId > 0 {
		lastDate = c.messages[msgId].Time
	}

	dateString := getDateString(lastDate, c.messages[msgId].Time)
	if dateString == "" {
		return 0
	}

	c.mntl.Return(1)
	c.mntl.Return(1)

	c.mntl.WriteAttributes(minigo.CaractereBleu)

	length := utf8.RuneCountInString(dateString)
	c.mntl.MoveCursorRight((minigo.ColonnesSimple - length) / 2)
	c.mntl.WriteString(dateString)

	c.mntl.WriteAttributes(minigo.CaractereBlanc)

	c.cache.MultBottom(-1, 2)
	return 2
}

func (c *ChatLayout) printMessage(msgId int) int {
	lines, vdt := FormatMessage(c.messages[msgId])
	c.mntl.Send(vdt)

	c.cache.MultBottom(msgId, lines)

	return lines
}

func (c *ChatLayout) Init() {
	c.mntl.CursorOff()

	// Load the last messages from DB
	if !c.getLastMessages() {
		return
	}

	c.mntl.MoveCursorAt(0, 1)

	for msgId := c.maxId; msgId < len(c.messages); msgId += 1 {
		curLine += c.printDate(msgId)
		curLine += c.printMessage(msgId)
	}
	c.maxId = len(c.messages) - 1

	// Return as much as possible to let empty space
	if curLine > 24 {
		curLine = 24
	}
	for ; curLine <= c.inputLine-2; curLine -= 1 {
		c.mntl.Return(1)
		c.cache.Bottom(-1)
	}

	c.printHelpers()
}
