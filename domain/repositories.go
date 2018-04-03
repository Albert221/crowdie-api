package domain

import (
	"gopkg.in/mgo.v2"
	"log"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"time"
)

type Repository interface {
	GetGroupById(id string) (*Group, error)
	UpdateMemberRole(id string, role int8) (*Member, error)
	UpdateMemberCoordsBit(id string, lat, lng float32, time time.Time) (*Member, error)
}

type MongoRepository struct {
	db *mgo.Database
}

func NewRepository(mongoURL, database string) *MongoRepository {
	s, err := mgo.Dial(mongoURL)
	if err != nil {
		log.Panicf("Cannot connect to '%s': %s", mongoURL, err)
	}

	s.SetMode(mgo.Monotonic, true)

	return &MongoRepository{
		db: s.DB(database),
	}
}

func (r MongoRepository) GetGroupById(id string) (*Group, error) {
	group := Group{}

	err := r.db.C("groups").Find(bson.M{"id": id}).One(&group)
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("group with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

func (r MongoRepository) UpdateMemberRole(id string, role int8) (*Member, error) {
	change := mgo.Change{
		Update: bson.M{"$set": bson.M{"members.$.role": role}},
		ReturnNew: true,
	}

	var updatedGroup Group
	_, err := r.db.C("groups").Find(bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}}).Apply(change, &updatedGroup)
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("member with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	for _, member := range updatedGroup.Members {
		if member.Id == id {
			return &member, nil
		}
	}

	return nil, fmt.Errorf("weird error, Id cannot be found")
}

func (r MongoRepository) UpdateMemberCoordsBit(id string, lat, lng float32, time time.Time) (*Member, error) {
	coordsBit := CoordsBit{
		Lat: lat,
		Lng: lng,
		Time: time,
	}

	change := mgo.Change{
		Update: bson.M{"$set": bson.M{"members.$.coordsbit": coordsBit}},
		ReturnNew: true,
	}

	var updatedGroup Group
	_, err := r.db.C("groups").Find(bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}}).Apply(change, &updatedGroup)
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("member with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	for _, member := range updatedGroup.Members {
		if member.Id == id {
			return &member, nil
		}
	}

	return nil, fmt.Errorf("weird error, Id cannot be found")
}