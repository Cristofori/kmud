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
	self.WriteLock()
	defer self.WriteUnlock()

	if name != self.Name {
		self.Name = name
		self.modified()
	}
}

func (self *Skill) GetEffect() types.SkillEffect {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Effect
}

func (self *Skill) SetEffect(effect types.SkillEffect) {
	self.WriteLock()
	defer self.WriteUnlock()

	if effect != self.Effect {
		self.Effect = effect
		self.modified()
	}
}

func (self *Skill) GetPower() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Power
}

func (self *Skill) SetPower(power int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if power != self.Power {
		self.Power = power
		self.modified()
	}
}

func (self *Skill) GetCost() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Cost
}

func (self *Skill) SetCost(cost int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if cost != self.Cost {
		self.Cost = cost
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

func (self *Skill) GetSpeed() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Speed
}

func (self *Skill) SetSpeed(speed int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if speed != self.Speed {
		self.Speed = speed
		self.modified()
	}
}
