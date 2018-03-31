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
	memberQuery := &bson.M{"members": &bson.M{"$elemMatch": &bson.M{"id": id}}}
	change := &bson.M{"members.$.role": role}

	err := r.db.C("groups").Update(memberQuery, &bson.M{"$set": &change})
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("member with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	// FIXME: Retrieve this member from query below. THIS ALWAYS RETURNS FIRST MEMBER!!!
	var member []Member
	r.db.C("groups").Find(&bson.M{"members.id": id}).Distinct("members.0", &member)

	return &member[0], nil
}

func (r MongoRepository) UpdateMemberCoordsBit(id string, lat, lng float32, time time.Time) (*Member, error) {
	memberQuery := &bson.M{"members": &bson.M{"$elemMatch": &bson.M{"id": id}}}

	coordsBit := CoordsBit{
		Lat: lat,
		Lng: lng,
		Time: time,
	}
	change := &bson.M{"members.$.coordsbit": coordsBit}

	err := r.db.C("groups").Update(memberQuery, &bson.M{"$set": &change})
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("member with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	// FIXME: Retrieve this member from query below. THIS ALWAYS RETURNS FIRST MEMBER!!!
	var member []Member
	r.db.C("groups").Find(&bson.M{"members.id": id}).Distinct("members.0", &member)

	return &member[0], nil
}