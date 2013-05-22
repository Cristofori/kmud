package database

type Zone struct {
	DbObject `bson:",inline"`
}

func NewZone(name string) *Zone {
	var zone Zone
	zone.initDbObject(name, zoneType)

	modified(&zone)
	return &zone
}

type Zones []*Zone

func (self Zones) Contains(z *Zone) bool {
	for _, zone := range self {
		if z == zone {
			return true
		}
	}

	return false
}

// vim: nocindent
