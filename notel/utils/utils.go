package utils

import (
	"time"
)

func WeekdayIdToStringShort(i time.Weekday) string {
	switch i {
	case time.Sunday:
		return "Dim."
	case time.Monday:
		return "Lun."
	case time.Tuesday:
		return "Mar."
	case time.Wednesday:
		return "Mer."
	case time.Thursday:
		return "Jeu."
	case time.Friday:
		return "Ven."
	case time.Saturday:
		return "Sam."
	}
	return ""
}

func WeekdayIdToString(i time.Weekday) string {
	switch i {
	case time.Sunday:
		return "Dimanche"
	case time.Monday:
		return "Lundi"
	case time.Tuesday:
		return "Mardi"
	case time.Wednesday:
		return "Mercredi"
	case time.Thursday:
		return "Jeudi"
	case time.Friday:
		return "Vendredi"
	case time.Saturday:
		return "Samedi"
	}
	return ""
}

func MonthIdToStringShort(i time.Month) string {
	switch i {
	case time.January:
		return "Jan."
	case time.February:
		return "Fév."
	case time.March:
		return "Mar."
	case time.April:
		return "Avr."
	case time.May:
		return "Mai"
	case time.June:
		return "Juin"
	case time.July:
		return "Jui."
	case time.August:
		return "Août"
	case time.September:
		return "Sep."
	case time.October:
		return "Oct."
	case time.November:
		return "Nov."
	case time.December:
		return "Déc."
	}
	return ""
}

func MonthIdToString(i time.Month) string {
	switch i {
	case time.January:
		return "Janvier"
	case time.February:
		return "Février"
	case time.March:
		return "Mars"
	case time.April:
		return "Avril"
	case time.May:
		return "Mai"
	case time.June:
		return "Juin"
	case time.July:
		return "Juillet"
	case time.August:
		return "Août"
	case time.September:
		return "Septembre"
	case time.October:
		return "Octobre"
	case time.November:
		return "Novembre"
	case time.December:
		return "Décembre"
	}
	return ""
}

func GetArrow(diff float64) string {
	if diff > 0 {
		return "↑"
	}
	return "↓"
}
