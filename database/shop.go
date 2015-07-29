package database

import (
	"github.com/Cristofori/kmud/types"
)

type Shop struct {
	DbObject `bson:",inline"`

	Inventory map[types.Id]ShopItem
}

type ShopItem struct {
	Price    int
	Quantity int
}

func NewShop() *Shop {
	var shop Shop
	return &shop
}
