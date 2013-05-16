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

// vim: nocindent
