package game

import (
	"fmt"
	"os"
	"strings"
)

func readMap(name string) {
	file, err := os.Open(name)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	stats, err := os.Stat(name)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	data := make([]byte, stats.Size())
	file.Read(data)
	fmt.Println(string(data))
}

func maps() []string {
	file, err := os.Open(".")

	var maps []string

	if err != nil {
		fmt.Println("Error:", err)
		return maps
	}

	names, err := file.Readdirnames(-1)

	if err != nil {
		fmt.Println("Error:", err)
		return maps
	}

	for _, filename := range names {
		if strings.Contains(filename, ".map") {
			maps = append(maps, filename)
		}
	}

	return maps
}

// vim: nocindent
