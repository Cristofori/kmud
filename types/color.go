package types

import (
	"fmt"
	"regexp"
	"strings"
)

var ColorRegex = regexp.MustCompile("([@#][0-6]|@@|##)")

type ColorMode int

const (
	ColorModeLight ColorMode = iota
	ColorModeDark  ColorMode = iota
	ColorModeNone  ColorMode = iota
)

type Color string

const (
	ColorRed     Color = "@0"
	ColorGreen   Color = "@1"
	ColorYellow  Color = "@2"
	ColorBlue    Color = "@3"
	ColorMagenta Color = "@4"
	ColorCyan    Color = "@5"
	ColorWhite   Color = "@6"

	ColorDarkRed     Color = "#0"
	ColorDarkGreen   Color = "#1"
	ColorDarkYellow  Color = "#2"
	ColorDarkBlue    Color = "#3"
	ColorDarkMagenta Color = "#4"
	ColorDarkCyan    Color = "#5"
	ColorBlack       Color = "#6"

	ColorGray   Color = "@@"
	ColorNormal Color = "##"
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

	gray   colorCode = "\033[22;37m"
	normal colorCode = "\033[0m"
)

func getAnsiCode(mode ColorMode, color Color) string {
	if mode == ColorModeNone {
		return ""
	}

	var code colorCode
	switch color {
	case ColorNormal:
		code = normal
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
	case ColorGray:
		code = gray
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
func Colorize(color Color, text string) string {
	return fmt.Sprintf("%s%s%s", string(color), text, string(ColorNormal))
}

var Lookup = map[Color]bool{
	ColorRed:         true,
	ColorGreen:       true,
	ColorYellow:      true,
	ColorBlue:        true,
	ColorMagenta:     true,
	ColorCyan:        true,
	ColorWhite:       true,
	ColorDarkRed:     true,
	ColorDarkGreen:   true,
	ColorDarkYellow:  true,
	ColorDarkBlue:    true,
	ColorDarkMagenta: true,
	ColorDarkCyan:    true,
	ColorBlack:       true,
	ColorGray:        true,
	ColorNormal:      true,
}

// Strips MUD color codes and replaces them with ansi color codes
func ProcessColors(text string, cm ColorMode) string {
	replace := func(match string) string {
		found := Lookup[Color(match)]

		if found {
			return getAnsiCode(cm, Color(match))
		}

		return match
	}

	after := ColorRegex.ReplaceAllStringFunc(text, replace)
	return after
}

func StripColors(text string) string {
	return ColorRegex.ReplaceAllString(text, "")
}
