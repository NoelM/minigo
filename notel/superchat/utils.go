package superchat

import (
	"fmt"
	"time"

	"github.com/NoelM/minigo"
	"github.com/NoelM/minigo/notel/databases"
	"github.com/NoelM/minigo/notel/utils"
)

func GetDateString(lastMsg, newMsg time.Time) (str string) {
	sinceLastMsg := newMsg.Sub(lastMsg)
	sinceToday := time.Since(lastMsg)

	location, _ := time.LoadLocation("Europe/Paris")
	newMsg = newMsg.In(location)

	if sinceLastMsg >= 365*24*time.Hour || (sinceToday > 24*time.Hour && sinceLastMsg > 30*time.Hour) {
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

func FormatMessage(msg databases.Message, dir RouleauDir, csi bool) (lines int, vdt [][]byte) {
	// Message Format
	// [nick]_[msg]
	formated := msg.Nick + " " + msg.Text

	// Wraps the message to 40 chars
	wrapped := minigo.WrapperLargeurNormale(formated)

	// If the direction is UP (up direction top of screen)
	// the first line to be printed is the last one
	var lineId int
	if dir == Up {
		lineId = len(wrapped) - 1
	}

	for {
		lineMsg := wrapped[lineId]

		buf := make([]byte, 0)
		if dir == Up {
			buf = append(buf, minigo.ReturnUp(1, csi)...)
		} else if dir == Down {
			buf = append(buf, minigo.Return(1, csi)...)
		}

		if lineId == 0 {
			buf = append(buf, minigo.EncodeAttributes(minigo.CaractereRouge)...)
			buf = append(buf, minigo.EncodeString(lineMsg[:len(msg.Nick)])...)
			buf = append(buf, minigo.EncodeAttribute(minigo.CaractereBlanc)...)

			buf = append(buf, minigo.EncodeString(lineMsg[len(msg.Nick):])...)
		} else {
			buf = append(buf, minigo.EncodeString(lineMsg)...)
		}

		vdt = append(vdt, buf)

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
