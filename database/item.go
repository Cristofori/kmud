package database

import (
	"github.com/Cristofori/kmud/types"
	"github.com/Cristofori/kmud/utils"
)

type Template struct {
	DbObject `bson:",inline"`
	Name     string
	Value    int
}

type Item struct {
	DbObject   `bson:",inline"`
	TemplateId types.Id
}

func NewTemplate(name string) *Template {
	template := &Template{
		Name: utils.FormatName(name),
	}
	template.init(template)
	return template
}

func NewItem(templateId types.Id) *Item {
	item := &Item{
		TemplateId: templateId,
	}
	item.init(item)
	return item
}

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
