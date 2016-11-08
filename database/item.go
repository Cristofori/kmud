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
	if name != self.GetName() {
		self.WriteLock()
		self.Name = utils.FormatName(name)
		self.WriteUnlock()
		self.modified()
	}
}

func (self *Template) SetValue(value int) {
	if value != self.GetValue() {
		self.WriteLock()
		self.Value = value
		self.WriteUnlock()
		self.modified()
	}
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
	if weight != self.GetWeight() {
		self.WriteLock()
		self.Weight = weight
		self.WriteUnlock()
		self.modified()
	}
}

func (self *Template) GetCapacity() int {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.Capacity
}

func (self *Template) SetCapacity(capacity int) {
	if capacity != self.GetCapacity() {
		self.WriteLock()
		self.Capacity = capacity
		self.WriteUnlock()
		self.modified()
	}
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
	if locked != self.IsLocked() {
		self.WriteLock()
		self.Locked = locked
		self.WriteUnlock()
		self.modified()
	}
}

func (self *Item) GetContainerId() types.Id {
	self.ReadLock()
	defer self.ReadUnlock()
	return self.ContainerId
}

func (self *Item) SetContainerId(id types.Id) bool {
	self.WriteLock()
	defer self.WriteUnlock()
	if id != self.ContainerId {
		self.ContainerId = id
		self.modified()
		return true
	}
	return false
}
