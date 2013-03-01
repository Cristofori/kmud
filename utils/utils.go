package utils

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"unicode"
)

func Write(conn io.Writer, text string) (int, error) {
	return WriteRaw(conn, []byte(text))
}

func WriteRaw(conn io.Writer, bytes []byte) (int, error) {
	return conn.Write(bytes)
}

func WriteLine(conn io.Writer, line string) (int, error) {
	return Write(conn, line+"\r\n")
}

// ClearLine sends the VT100 code for erasing the line followed by a carriage
// return to move the cursor back to the beginning of the line
func ClearLine(conn io.Writer) {
	clearline := "\x1B[2K"
	Write(conn, clearline+"\r")
}

func Simplify(str string) string {
	simpleStr := strings.TrimSpace(str)
	simpleStr = strings.ToLower(simpleStr)
	return simpleStr
}

func GetRawUserInput(conn io.ReadWriter, prompt string) string {
	reader := bufio.NewReader(conn)

	for {
		Write(conn, prompt)

		input, err := reader.ReadString('\n')
		input = strings.Trim(input, "\r\n")

		PanicIfError(err)

		if input == "x" || input == "X" {
			return ""
		} else if input != "" {
			return input
		}
	}

	panic("Unexpected code path")
	return ""
}

func GetUserInput(conn io.ReadWriter, prompt string) string {
	input := GetRawUserInput(conn, prompt)
	return Simplify(input)
}

func HandleError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func FormatName(name string) string {
	if name == "" {
		return name
	}

	runes := []rune(Simplify(name))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func Argify(data string) (string, []string) {
	fields := strings.Fields(data)

	if len(fields) == 0 {
		return "", []string{}
	}

	arg1 := Simplify(fields[0])
	args := fields[1:]

	return arg1, args
}

func rowEmpty(row string) bool {
	for _, char := range row {
		if char != ' ' {
			return false
		}
	}
	return true
}

func TrimUpperRows(rows []string) []string {
	for _, row := range rows {
		if !rowEmpty(row) {
			break
		}

		rows = rows[1:]
	}

	return rows
}

func TrimLowerRows(rows []string) []string {
	for i := len(rows) - 1; i >= 0; i -= 1 {
		row := rows[i]
		if !rowEmpty(row) {
			break
		}
		rows = rows[:len(rows)-1]
	}

	return rows
}

func TrimEmptyRows(str string) string {
	rows := strings.Split(str, "\n")
	return strings.Join(TrimLowerRows(TrimUpperRows(rows)), "\n")
}

func ValidateName(name string) error {
	const MinSize = 3
	const MaxSize = 12

	if len(name) < MinSize || len(name) > MaxSize {
		return errors.New(fmt.Sprintf("Names must be between %v and %v letters long", MinSize, MaxSize))
	}

	regex := regexp.MustCompile("^[a-zA-Z]*$")

	if !regex.MatchString(name) {
		return errors.New("Names may only contain letters (A-Z)")
	}

	return nil
}

func MonitorChannel() {
	// TODO: See if there's a way to take in a generic channel and see how close it is to being full 
}

// BestMatch searches the given list for the given pattern, the index of the
// longest match that starts with the given pattern is returned. Returns -1 if
// no match was found, -2 if the result is ambiguous. The search is case
// insensitive
func BestMatch(pattern string, searchList []string) int {
	pattern = strings.ToLower(pattern)

	index := -1

	for i, searchItem := range searchList {
		searchItem = strings.ToLower(searchItem)

		if searchItem == pattern {
			return i
		}

		if strings.HasPrefix(searchItem, pattern) {
			if index != -1 {
				return -2
			}

			index = i
		}
	}

	return index
}

func compress(data []byte) []byte {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

type WatchableReadWriter struct {
	rw       io.ReadWriter
	watchers []io.ReadWriter
}

func NewWatchableReadWriter(rw io.ReadWriter) *WatchableReadWriter {
	var watchable WatchableReadWriter
	watchable.rw = rw
	return &watchable
}

func (w *WatchableReadWriter) Read(p []byte) (int, error) {
	n, err := w.rw.Read(p)

	for _, watcher := range w.watchers {
		watcher.Write(p[:n])
	}

	return n, err
}

func (w *WatchableReadWriter) Write(p []byte) (int, error) {
	for _, watcher := range w.watchers {
		watcher.Write(p)
	}

	return w.rw.Write(p)
}

func (w *WatchableReadWriter) AddWatcher(rw io.ReadWriter) {
	w.watchers = append(w.watchers, rw)
}

func (w *WatchableReadWriter) RemoveWatcher(rw io.ReadWriter) {
	for i, watcher := range w.watchers {
		if watcher == rw {
			// TODO: Potential memory leak. See http://code.google.com/p/go-wiki/wiki/SliceTricks
			w.watchers = append(w.watchers[:i], w.watchers[i+1:]...)
			return
		}
	}
}

// vim: nocindent
