package domain

import (
	"fmt"
	"time"
)

var groups []Group

func SetupMockData() {
	groups = append(groups, Group{
		Id: "d175a80a-399a-4c89-b05a-1b8e2decab57",
		Members: []Member{
			{
				Id:   "4f8b6c3c-a1d0-44a7-9020-2ef3b87e6fc7",
				Name: "Albert",
				Role: 1,
			  	CoordsBits: []CoordsBit{
			  		{
						Lat: 54.522117,
						Lng: 18.530506,
			  			Time: time.Now().Add(-5 * time.Second),
					},
				},
			},
			{
				Id: "eeb30046-6275-4534-ad60-427726cbaeb7",
				Name: "Jan",
				Role: 0,
				CoordsBits: []CoordsBit{
					{
						Lat: 54.522026,
						Lng: 18.532108,
						Time: time.Now().Add(-3 * time.Second),
					},
				},
			},
			{
				Id: "dd8f54e9-cff4-4035-955f-4d16e9a6a714",
				Name: "Kuba",
				Role: 0,
				CoordsBits: []CoordsBit{
					{
						Lat: 54.522194,
						Lng: 18.530834,
						Time: time.Now(),
					},
				},
			},
		},
	})
}

func GetGroupById(id string) (*Group, error) {
	for _, group := range groups {
		if group.Id == id {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("group with ID '%s' does not exist", id)
}