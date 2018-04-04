package domain

import "time"

const (
	USER = 0
	ADMIN = 1
)

type Group struct {
	Id      string   `json:"id"`
	Members []Member `json:"members"`
}

type Member struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Role      int8      `json:"role"`
	CoordsBit CoordsBit `json:"coordsBit"`
}

type CoordsBit struct {
	Lat  float32   `json:"lat"`
	Lng  float32   `json:"lng"`
	Time time.Time `json:"time"`
}
