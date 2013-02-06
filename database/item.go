package database

type Item struct {
	DbObject `bson:",inline"`
}

func NewItem(name string) *Item {
	var item Item
	item.initDbObject(name, itemType)

	return &item
}

// vim: nocindent
