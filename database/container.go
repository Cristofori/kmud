package database

type Container struct {
	DbObject `bson:",inline"`
	Cash     int
	Capacity int
	Weight   int
}

func (self *Container) SetCash(cash int) {
	self.WriteLock()
	defer self.WriteUnlock()

	if cash != self.Cash {
		self.Cash = cash
		self.modified()
	}
}

func (self *Container) GetCash() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Cash
}

func (self *Container) AddCash(amount int) {
	self.SetCash(self.GetCash() + amount)
}

func (self *Container) RemoveCash(amount int) {
	self.SetCash(self.GetCash() - amount)
}

func (self *Container) GetCapacity() int {
	self.ReadLock()
	defer self.ReadUnlock()

	return self.Capacity
}

func (self *Container) SetCapacity(limit int) {
	if limit != self.GetCapacity() {
		self.WriteLock()
		self.Capacity = limit
		self.WriteUnlock()
		self.modified()
	}
}
