package utils

import (
	"fmt"
	"strings"
    "regexp"
)

type ColorMode int

const (
	ColorModeLight ColorMode = iota
	ColorModeDark  ColorMode = iota
	ColorModeNone  ColorMode = iota
)

type Color string

const (
	ColorRed         Color = "@0"
	ColorGreen       Color = "@1"
	ColorYellow      Color = "@2"
	ColorBlue        Color = "@3"
	ColorMagenta     Color = "@4"
	ColorCyan        Color = "@5"
	ColorWhite       Color = "@6"

	ColorDarkRed     Color = "#0"
	ColorDarkGreen   Color = "#1"
	ColorDarkYellow  Color = "#2"
	ColorDarkBlue    Color = "#3"
	ColorDarkMagenta Color = "#4"
	ColorDarkCyan    Color = "#5"
	ColorBlack       Color = "#6"

	ColorGray        Color = "@@"
    ColorNormal      Color = "##"
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
    normal colorCode = "\033[0m"
)

func getAsciiCode(mode ColorMode, color Color) string {
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
	return fmt.Sprintf("%s%s%s", string(color), text, string(ColorNormal))
}

// Strips MUD color codes and replaces them with ascii color codes
func processColors(text string, cm ColorMode) string {
    regex := regexp.MustCompile("([@#][0-6]|@@|##)")

    lookup := map[Color]bool{}

    lookup[ColorRed] = true
    lookup[ColorGreen] = true
    lookup[ColorYellow] = true
    lookup[ColorBlue] = true
    lookup[ColorMagenta] = true
    lookup[ColorCyan] = true
    lookup[ColorWhite] = true

    lookup[ColorDarkRed] = true
    lookup[ColorDarkGreen] = true
    lookup[ColorDarkYellow] = true
    lookup[ColorDarkBlue] = true
    lookup[ColorDarkMagenta] = true
    lookup[ColorDarkCyan] = true
    lookup[ColorBlack] = true

    lookup[ColorGray] = true
    lookup[ColorNormal] = true

    replace := func(match string) string {
        _, found := lookup[Color(match)]

        if found {
            return getAsciiCode(cm, Color(match))
        }

        return match
    }

    after := regex.ReplaceAllStringFunc(text, replace)
    return after
}

// vim: nocindent
