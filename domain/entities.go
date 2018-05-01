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

type (
	Group struct {
		Id      string   `json:"id"`
		Members []Member `json:"members"`
	}

	Member struct {
		Id        string    `json:"id"`
		Name      string    `json:"name"`
		Role      int8      `json:"role"`
		AndroidId string    `json:"androidId,omitempty"`
		Secret    string    `json:"secret,omitempty"`
		CoordsBit CoordsBit `json:"coordsBit"`
	}

	CoordsBit struct {
		Lat  float32   `json:"lat"`
		Lng  float32   `json:"lng"`
		Time time.Time `json:"time"`
	}
)

func NewGroup(creator Member) Group {
	return Group{
		Id: uuid.New().String(),
		Members: []Member{
			creator,
		},
	}
}

// Export returns a copy of given group but without sensitive data.
func (g Group) Export() Group {
	group := g
	for i := range group.Members {
		group.Members[i].Secret = ""
		group.Members[i].AndroidId = ""
	}

	return group
}

func NewMember(name, androidId string, lat, lng float32) Member {
	// TODO: Generate secure random string
	secret := ""

	return Member{
		Id:        uuid.New().String(),
		Name:      name,
		Role:      USER,
		AndroidId: androidId,
		Secret:    secret,
		CoordsBit: CoordsBit{
			Lat:  lat,
			Lng:  lng,
			Time: time.Now(),
		},
	}
}

var (
	GroupNotExists = errors.New("group does not exist")

	MemberNotExists         = errors.New("member does not exist")
	NoSufficientPermissions = errors.New("member does not have sufficient permissions")
)
