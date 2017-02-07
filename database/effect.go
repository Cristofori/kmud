package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Effect struct {
	DbObject `bson:",inline"`

	Type     types.EffectKind
	Name     string
	Power    int
	Cost     int
	Variance int
	Speed    int
	Time     int
}

func NewEffect(name string) types.Effect {
	effect := &Effect{
		Name:     utils.FormatName(name),
		Power:    1,
		Type:     types.HitpointEffect,
		Variance: 0,
		Time:     1,
	}

	dbinit(effect)
	return effect
}

func (self *Effect) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Name
}

func (self *Effect) SetName(name string) {
	self.writeLock(func() {
		self.Name = utils.FormatName(name)
	})
}

func (self *Effect) GetType() types.EffectKind {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Type
}

func (self *Effect) SetType(effectKind types.EffectKind) {
	self.writeLock(func() {
		self.Type = effectKind
	})
}

func (self *Effect) GetPower() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Power
}

func (self *Effect) SetPower(power int) {
	self.writeLock(func() {
		self.Power = power
	})
}

func (self *Effect) GetCost() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Cost
}

func (self *Effect) SetCost(cost int) {
	self.writeLock(func() {
		self.Cost = cost
	})
}

func (self *Effect) GetVariance() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Variance
}

func (self *Effect) SetVariance(variance int) {
	self.writeLock(func() {
		self.Variance = variance
	})
}

func (self *Effect) GetSpeed() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Speed
}

func (self *Effect) SetSpeed(speed int) {
	self.writeLock(func() {
		self.Speed = speed
	})
}

func (self *Effect) GetTime() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Time
}

func (self *Effect) SetTime(speed int) {
	self.writeLock(func() {
		self.Time = speed
	})
}
