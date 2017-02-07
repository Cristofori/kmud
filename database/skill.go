package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Skill struct {
	DbObject `bson:",inline"`

	Effects utils.Set
	Name    string
}

func NewSkill(name string) *Skill {
	skill := &Skill{
		Name: utils.FormatName(name),
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
		self.Name = utils.FormatName(name)
	})
}

func (self *Skill) AddEffect(id types.Id) {
	self.writeLock(func() {
		if self.Effects == nil {
			self.Effects = utils.Set{}
		}
		self.Effects.Insert(id.Hex())
	})
}

func (self *Skill) RemoveEffect(id types.Id) {
	self.writeLock(func() {
		self.Effects.Remove(id.Hex())
	})
}

func (self *Skill) GetEffects() []types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return idSetToList(self.Effects)
}

func (self *Skill) HasEffect(id types.Id) bool {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Effects.Contains(id.Hex())
}
