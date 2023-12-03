package minichat

import (
	"time"
	"unicode/utf8"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
)

type ChatLayout struct {
	mntl *minigo.Minitel

	msgDB    *databases.MessageDatabase
	messages []databases.Message

	nick string

	inputLine   int
	inputHeight int

	minId int
	maxId int
}

func NewChatLayout(mntl *minigo.Minitel, msgDB *databases.MessageDatabase, nick string, inputLine, inputHeight int) *ChatLayout {
	return &ChatLayout{
		mntl:        mntl,
		msgDB:       msgDB,
		nick:        nick,
		inputLine:   InputLine,
		inputHeight: inputHeight,
		minId:       -1,
		maxId:       -1,
	}
}

func (c *ChatLayout) cleanHelpers() {
	c.mntl.MoveCursorAt(c.inputLine-1, 1)
	c.mntl.CleanLine()

	c.mntl.MoveCursorAt(24, 1)
	c.mntl.CleanLine()
}

func (c *ChatLayout) printHelpers() {
	c.mntl.HLine(c.inputLine-1, 1, 40, minigo.HCenter)

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

func (c *ChatLayout) printDate(lastDate, curDate time.Time) int {
	dateString := getDateString(lastDate, curDate)
	if dateString == "" {
		return 0
	}

	c.mntl.Return(1)
	c.mntl.WriteAttributes(minigo.CaractereBleu)

	length := utf8.RuneCountInString(dateString)
	c.mntl.MoveCursorRight((minigo.ColonnesSimple - length) / 2)
	c.mntl.WriteString(dateString)

	c.mntl.WriteAttributes(minigo.CaractereBlanc)
	c.mntl.Return(1)

	return 2
}

func (c *ChatLayout) printMessage(msg databases.Message) int {
	lines, vdt := FormatMessage(msg)
	c.mntl.Send(vdt)

	return lines
}

func (c *ChatLayout) Full() {
	c.mntl.CursorOff()
	c.cleanHelpers()

	// Load the last messages from DB
	if !c.getLastMessages() {
		return
	}
	if len(c.messages) == 0 {
		return
	}

	// LINE        | CONTENT
	// ------------|------------------------
	// 1           | Message Zone
	// ...         | ...
	// InputLine-2 | Last message line
	// InputLine-1 | HLine <-- restart HERE
	// InputLine   | Input
	curLine := c.inputLine - 1
	c.mntl.MoveCursorAt(curLine, 1)

	// Not message displayed yet
	if c.maxId < 0 {
		if c.maxId = len(c.messages) - 10; c.maxId < 0 {
			c.maxId = 0
		}
	}

	var lastDate time.Time

	for _, msg := range c.messages[c.maxId:] {
		curLine += c.printDate(lastDate, msg.Time)
		curLine += c.printMessage(msg)

		lastDate = msg.Time
	}
	c.maxId = len(c.messages) - 1

	// Return as much as possible to let empty space
	if curLine > 24 {
		curLine = 24
	}
	for ; curLine <= c.inputLine-2; curLine -= 1 {
		c.mntl.Return(1)
	}

	c.printHelpers()
}
