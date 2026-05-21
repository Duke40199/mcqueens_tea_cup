package entity

import "time"

type StoreLocation struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	CountryCode  string    `json:"country_code"`
	SegaAreaCode string    `json:"sega_area_code"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s StoreLocation) TableName() string {
	return "all_net_store_locations"
}
