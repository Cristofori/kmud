package database

import (
	"fmt"
	"time"

	"github.com/Cristofori/kmud/types"
)

type World struct {
	DbObject `bson:",inline"`
	Time     time.Time
}

func NewWorld() *World {
	world := &World{Time: time.Now()}

	world.init(world)
	return world
}

type _time struct {
	hour int
	min  int
	sec  int
}

func (self _time) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", self.hour, self.min, self.sec)
}

const _TIME_MULTIPLIER = 3

func (self *World) GetTime() types.Time {
	self.ReadLock()
	defer self.ReadUnlock()

	hour, min, sec := self.Time.Clock()
	return _time{hour: hour, min: min, sec: sec}
}

func (self *World) AdvanceTime() {
	self.WriteLock()
	defer self.WriteUnlock()

	self.Time = self.Time.Add(3 * time.Second)
}
