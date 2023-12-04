package minichat

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/utils"
)

func GetDateString(lastMsg, newMsg time.Time) (str string) {
	sinceLastMsg := newMsg.Sub(lastMsg)

	if sinceLastMsg >= 365*24*time.Hour {
		// More than a YEAR
		//
		// Lun. 24 Juin 2023, 14:30
		str = fmt.Sprintf("%s %d %s %d, %s",
			utils.WeekdayIdToString(newMsg.Weekday()),
			newMsg.Day(),
			utils.MonthIdToString(newMsg.Month()),
			newMsg.Year(),
			newMsg.Format("15:04"))

	} else if sinceLastMsg >= 24*time.Hour || newMsg.Day() != lastMsg.Day() {
		// More than 24 HOURS
		// Or if the date changed (like a message between 23:59 and 00:00)
		//
		// Lun. 24 Juin, 14:30
		str = fmt.Sprintf("%s %d %s, %s",
			utils.WeekdayIdToString(newMsg.Weekday()),
			newMsg.Day(),
			utils.MonthIdToString(newMsg.Month()),
			newMsg.Format("15:04"))

	} else if sinceLastMsg > 10*time.Minute {
		// More than 10 MINUTES
		//
		// 14:30
		str = newMsg.Format("15:04")
	}

	return
}

func FormatMessage(msg databases.Message, dir RouleauDir) (lines int, vdt []byte) {
	// Message Format
	// [nick]_[msg]
	formated := msg.Nick + " " + msg.Text

	// Wraps the message to 40 chars
	wrapped := minigo.WrapperLargeurNormale(formated)

	var lineId int
	if dir == Up {
		lineId = len(wrapped)
	}

	for {
		lineMsg := wrapped[lineId]

		if dir == Up {
			vdt = append(vdt, minigo.GetMoveCursorUp(1)...)
		} else if dir == Down {
			vdt = append(vdt, minigo.GetMoveCursorReturn(1)...)
		}

		if lineId == 0 {
			vdt = append(vdt, minigo.EncodeAttributes(minigo.CaractereRouge)...)
			vdt = append(vdt, minigo.EncodeString(lineMsg[:len(msg.Nick)])...)
			vdt = append(vdt, minigo.EncodeAttribute(minigo.CaractereBlanc)...)

			vdt = append(vdt, minigo.EncodeString(lineMsg[len(msg.Nick):])...)
		} else {
			vdt = append(vdt, minigo.EncodeString(lineMsg)...)
		}

		if dir == Up {
			if lineId -= 1; lineId < 0 {
				break
			}

		} else if dir == Down {
			if lineId += 1; lineId == len(wrapped) {
				break
			}
		}
	}

	return len(wrapped), vdt
}
