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

func FormatMessage(msg databases.Message) (lines int, vdt []byte) {
	// Message Format
	// [nick]_[msg]
	formated := msg.Nick + " " + msg.Text
	wrapped := minigo.WrapperLargeurNormale(formated)

	// One line of 40 runes max
	for lineId, lineMsg := range wrapped {
		if lineId == 0 {
			vdt = append(vdt, minigo.EncodeAttributes(minigo.CaractereRouge)...)
			vdt = append(vdt, minigo.EncodeString(lineMsg[:len(msg.Nick)])...)
			vdt = append(vdt, minigo.EncodeAttribute(minigo.CaractereBlanc)...)

			vdt = append(vdt, minigo.EncodeString(lineMsg[len(msg.Nick):])...)
		} else {
			vdt = append(vdt, minigo.EncodeString(lineMsg)...)
		}

		vdt = append(vdt, minigo.GetMoveCursorDown(1)...)
	}

	return len(wrapped), vdt
}
