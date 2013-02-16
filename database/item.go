package database

type Item struct {
	DbObject `bson:",inline"`
}

func NewItem(name string) *Item {
	var item Item
	item.initDbObject(name, itemType)

	return &item
}

func ItemNames(items []*Item) []string {
	names := make([]string, len(items))

	for i, item := range items {
		names[i] = item.PrettyName()
	}

	return names
}

// vim: nocindent
