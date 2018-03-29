package domain

import "time"

type Group struct {
	Id      string   `json:"id"`
	Members []Member `json:"members"`
}

type Member struct {
	Id         string      `json:"id"`
	Name       string      `json:"name"`
	Role       int8        `json:"role"`
	CoordsBits []CoordsBit `json:"coordsBits"`
}

type CoordsBit struct {
	Lat  float32   `json:"lat"`
	Lng  float32   `json:"lng"`
	Time time.Time `json:"time"`
}

func (m Member) GetLatestCoords() *CoordsBit {
	var latestBit *CoordsBit

	for _, bit := range m.CoordsBits {
		if latestBit == nil || bit.Time.After(latestBit.Time) {
			latestBit = &bit
		}
	}

	return latestBit
}
