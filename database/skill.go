package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Skill struct {
	DbObject `bson:",inline"`

	Name     string
	Power    int
	Cost     int
	Variance int
	Speed    int
	Effect   types.SkillEffect
}

func NewSkill(name string, power int) *Skill {
	skill := &Skill{
		Name:     utils.FormatName(name),
		Power:    power,
		Effect:   types.DamageEffect,
		Variance: 0,
	}

	dbinit(skill)
	return skill
}

func (self *Skill) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Skill) SetName(name string) {
	self.writeLock(func() {
		self.Name = name
	})
}

func (self *Skill) GetEffect() types.SkillEffect {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Effect
}

func (self *Skill) SetEffect(effect types.SkillEffect) {
	self.writeLock(func() {
		self.Effect = effect
	})
}

func (self *Skill) GetPower() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Power
}

func (self *Skill) SetPower(power int) {
	self.writeLock(func() {
		self.Power = power
	})
}

func (self *Skill) GetCost() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Cost
}

func (self *Skill) SetCost(cost int) {
	self.writeLock(func() {
		self.Cost = cost
	})
}

func (self *Skill) GetVariance() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Variance
}

func (self *Skill) SetVariance(variance int) {
	self.writeLock(func() {
		self.Variance = variance
	})
}

func (self *Skill) GetSpeed() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Speed
}

func (self *Skill) SetSpeed(speed int) {
	self.writeLock(func() {
		self.Speed = speed
	})
}
