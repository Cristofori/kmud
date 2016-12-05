package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Template struct {
	DbObject `bson:",inline"`
	Name     string
	Value    int
	Weight   int
	Capacity int
}

type Item struct {
	Container `bson:",inline"`

	TemplateId  types.Id
	Locked      bool
	ContainerId types.Id
}

func NewTemplate(name string) *Template {
	template := &Template{
		Name: utils.FormatName(name),
	}
	dbinit(template)
	return template
}

func NewItem(templateId types.Id) *Item {
	item := &Item{
		TemplateId: templateId,
	}
	dbinit(item)
	return item
}

// Template

func (self *Template) GetName() string {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Name
}

func (self *Template) SetName(name string) {
	self.writeLock(func() {
		self.Name = utils.FormatName(name)
	})
}

func (self *Template) SetValue(value int) {
	self.writeLock(func() {
		self.Value = value
	})
}

func (self *Template) GetValue() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Value
}

func (self *Template) GetWeight() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Weight
}

func (self *Template) SetWeight(weight int) {
	self.writeLock(func() {
		self.Weight = weight
	})
}

func (self *Template) GetCapacity() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Capacity
}

func (self *Template) SetCapacity(capacity int) {
	self.writeLock(func() {
		self.Capacity = capacity
	})
}

// Item

func (self *Item) GetTemplateId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.TemplateId
}

func (self *Item) GetTemplate() types.Template {
	self.ReadLock()
	defer self.ReadUnlock()
	return Retrieve(self.TemplateId, types.TemplateType).(types.Template)
}

func (self *Item) GetName() string {
	return self.GetTemplate().GetName()
}

func (self *Item) GetValue() int {
	return self.GetTemplate().GetValue()
}

func (self *Item) GetCapacity() int {
	return self.GetTemplate().GetCapacity()
}

func (self *Item) IsLocked() bool {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Locked
}

func (self *Item) SetLocked(locked bool) {
	self.writeLock(func() {
		self.Locked = locked
	})
}

func (self *Item) GetContainerId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.ContainerId
}

func (self *Item) SetContainerId(id types.Id, from types.Id) bool {
	self.WriteLock()
	if from != self.ContainerId {
		self.WriteUnlock()
		return false
	}
	self.ContainerId = id
	self.WriteUnlock()
	self.syncModified()
	return true
}
