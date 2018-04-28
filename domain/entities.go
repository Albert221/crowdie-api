package domain

import (
	"time"
	"errors"
	"github.com/google/uuid"
)

const (
	USER  = 0
	ADMIN = 1
)

type Group struct {
	Id      string   `json:"id"`
	Members []Member `json:"members"`
}

func NewGroup(creator Member) Group {
	return Group{
		Id: uuid.New().String(),
		Members: []Member{
			creator,
		},
	}
}

type Member struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Role      int8      `json:"role"`
	AndroidId string    `json:"androidId"`
	CoordsBit CoordsBit `json:"coordsBit"`
}

func NewMember(name, androidId string, lat, lng float32) Member {
	return Member{
		Id:        uuid.New().String(),
		Name:      name,
		Role:      USER,
		AndroidId: androidId,
		CoordsBit: CoordsBit{
			Lat:  lat,
			Lng:  lng,
			Time: time.Now(),
		},
	}
}

type CoordsBit struct {
	Lat  float32   `json:"lat"`
	Lng  float32   `json:"lng"`
	Time time.Time `json:"time"`
}

var (
	GroupNotExists = errors.New("group does not exist")

	MemberNotExists = errors.New("member does not exist")
)
