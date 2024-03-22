package minigo

import (
	"strings"
	"unicode/utf8"
)

func minInt(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func maxInt(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func WrapperLargeurNormale(text string) []string {
	return Wrapper(text, ColonnesSimple)
}

func WrapperLargeurDouble(text string) []string {
	return Wrapper(text, ColonnesDouble)
}

func Wrapper(text string, size int) (wrapped []string) {
	var words []string
	length := 0

	for _, s := range strings.Split(text, " ") {
		if length+utf8.RuneCountInString(s)+1 >= size {
			wrapped = append(wrapped, strings.Join(words, " "))

			length = 0
			words = []string{}
		}

		length += utf8.RuneCountInString(s) + 1 // the size of the space
		words = append(words, s)
	}

	if len(words) > 0 {
		wrapped = append(wrapped, strings.Join(words, " "))
	}

	return
}
