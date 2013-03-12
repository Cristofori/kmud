package utils

import (
	"strconv"
	"unicode"
)

// SortableStrings implements the sort.Interface for a []string. It uses "natural" sort
// order rather than asciibetical sort order.
type SortableStrings []string

func (s SortableStrings) Len() int {
	return len(s)
}

func (s SortableStrings) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortableStrings) Less(i, j int) bool {
	pieces1 := piecesOf(s[i])
	pieces2 := piecesOf(s[j])

	for i, piece1 := range pieces1 {
		if i >= len(pieces2) {
			return true
		}

		piece2 := pieces2[i]

		if piece1 != piece2 {
			if unicode.IsDigit(rune(piece1[0])) {
				num1, _ := strconv.Atoi(piece1)
				num2, _ := strconv.Atoi(piece2)
				return num1 < num2
			} else {
				return piece1 < piece2
			}
		}
	}

	return false
}

func piecesOf(str string) []string {
	var pieces []string

	if len(str) == 0 {
		return pieces
	}

	type Mode int
	const (
		CharMode = iota
		NumMode  = iota
	)

	currentMode := CharMode

	if unicode.IsDigit(rune(str[0])) {
		currentMode = NumMode
	}

	begin := 0

	for i, c := range str {
		newMode := CharMode
		if unicode.IsDigit(c) {
			newMode = NumMode
		}

		if newMode != currentMode {
			pieces = append(pieces, str[begin:i])
			begin = i
			currentMode = newMode
		}
	}

	pieces = append(pieces, str[begin:])

	return pieces
}

// vim: nocindent
