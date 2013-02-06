package database

type Zone struct {
	DbObject `bson:",inline"`
}

const ()

func NewZone(name string) *Zone {
	var zone Zone
	zone.initDbObject(zoneType)

	return &zone
}

// vim: nocindent
