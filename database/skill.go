package database

type Skill struct {
	DbObject `bson:",inline"`

	Name   string
	Damage int
}

func NewSkill(name string, damage int) *Skill {
	skill := Skill{
		Name:   name,
		Damage: damage,
	}

	skill.initDbObject(&skill)
	return &skill
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

func (self *Skill) SetDamage(damage int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if damage != self.Damage {
		self.Damage = damage
		self.modified()
	}
}

func (self *Skill) GetDamage() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Damage
}
