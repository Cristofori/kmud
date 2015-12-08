package utils

type Set map[string]bool

func (self Set) Contains(key string) bool {
	_, found := self[key]
	return found
}

func (self Set) Insert(key string) {
	self[key] = true
}

func (self Set) Remove(key string) {
	delete(self, key)
}

func (self Set) Size() int {
	return len(self)
}
