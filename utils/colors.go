package utils

import (
	"fmt"
	"strings"
)

type ColorMode int

const (
	ColorModeLight ColorMode = iota
	ColorModeDark  ColorMode = iota
	ColorModeNone  ColorMode = iota
)

type Color int

const ColorNormal string = "\033[0m"

const (
	ColorRed         Color = iota
	ColorGreen       Color = iota
	ColorYellow      Color = iota
	ColorBlue        Color = iota
	ColorMagenta     Color = iota
	ColorCyan        Color = iota
	ColorWhite       Color = iota
	ColorDarkRed     Color = iota
	ColorDarkGreen   Color = iota
	ColorDarkYellow  Color = iota
	ColorDarkBlue    Color = iota
	ColorDarkMagenta Color = iota
	ColorDarkCyan    Color = iota
	ColorBlack       Color = iota
	ColorGray        Color = iota
)

type colorCode string

const (
	red     colorCode = "\033[01;31m"
	green   colorCode = "\033[01;32m"
	yellow  colorCode = "\033[01;33m"
	blue    colorCode = "\033[01;34m"
	magenta colorCode = "\033[01;35m"
	cyan    colorCode = "\033[01;36m"
	white   colorCode = "\033[01;37m"

	darkRed     colorCode = "\033[22;31m"
	darkGreen   colorCode = "\033[22;32m"
	darkYellow  colorCode = "\033[22;33m"
	darkBlue    colorCode = "\033[22;34m"
	darkMagenta colorCode = "\033[22;35m"
	darkCyan    colorCode = "\033[22;36m"
	black       colorCode = "\033[22;30m"

	gray colorCode = "\033[22;37m"
)

func GetColor(mode ColorMode, color Color) string {
	if mode == ColorModeNone {
		return ""
	}

	var code colorCode
	switch color {
	case ColorRed:
		code = red
	case ColorGreen:
		code = green
	case ColorYellow:
		code = yellow
	case ColorBlue:
		code = blue
	case ColorMagenta:
		code = magenta
	case ColorCyan:
		code = cyan
	case ColorWhite:
		code = white
	case ColorDarkRed:
		code = darkRed
	case ColorDarkGreen:
		code = darkGreen
	case ColorDarkYellow:
		code = darkYellow
	case ColorDarkBlue:
		code = darkBlue
	case ColorDarkMagenta:
		code = darkMagenta
	case ColorDarkCyan:
		code = darkCyan
	case ColorBlack:
		code = black
	}

	if mode == ColorModeDark {
		if code == white {
			return string(black)
		} else if code == black {
			return string(white)
		} else if strings.Contains(string(code), "01") {
			return strings.Replace(string(code), "01", "22", 1)
		} else {
			return strings.Replace(string(code), "22", "01", 1)
		}
	}

	return string(code)
}

// Wraps the given text in the given color, followed by a color reset
func Colorize(cm ColorMode, color Color, text string) string {
	colorStr := GetColor(cm, color)

	after := ColorNormal
	if cm == ColorModeNone {
		after = ""
	}

	return fmt.Sprintf("%s%s%s", colorStr, text, after)
}

// vim: nocindent
