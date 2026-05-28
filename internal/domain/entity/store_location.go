package entity

import "time"

type StoreLocation struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	SegaAreaCode   string    `json:"sega_area_code"`
	AllNetAreaCode string    `json:"all_net_area_code"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s StoreLocation) TableName() string {
	return "idac_stores"
}
