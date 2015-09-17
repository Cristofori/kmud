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
	Effect   types.SkillEffect
}

func NewSkill(name string, power int) *Skill {
	skill := &Skill{
		Name:     utils.FormatName(name),
		Power:    power,
		Effect:   types.DamageEffect,
		Variance: 0,
	}

	skill.init(skill)
	return skill
}

func (self *Skill) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Skill) SetName(name string) {
	self.WriteLock()
	defer self.WriteUnlock()

	if name != self.Name {
		self.Name = name
		self.modified()
	}
}

func (self *Skill) GetPower() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Power
}

func (self *Skill) SetPower(damage int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if damage != self.Power {
		self.Power = damage
		self.modified()
	}
}

func (self *Skill) GetVariance() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Variance
}

func (self *Skill) SetVariance(variance int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if variance != self.Variance {
		self.Variance = variance
		self.modified()
	}
}
