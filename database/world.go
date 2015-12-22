package database

import (
	"fmt"
	"time"
)

type World struct {
	DbObject `bson:",inline"`
}

type Time struct {
	hour int
	min  int
	sec  int
}

func (self Time) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", self.hour, self.min, self.sec)
}

const _TIME_MULTIPLIER = 3

// Returns the time of day
func GetTime() Time {
	hour, min, sec := time.Now().Clock()

	const SecondsInADay = 60 * 60 * 24

	totalSeconds := sec + (min * 60) + (hour * 60 * 60)
	totalSeconds = totalSeconds * _TIME_MULTIPLIER

	hour = totalSeconds / (60 * 60)
	hour = hour % 24
	totalSeconds = totalSeconds % (60 * 60)

	min = totalSeconds / 60
	totalSeconds = totalSeconds % 60

	sec = totalSeconds

	return Time{hour: hour, min: min, sec: sec}
}
