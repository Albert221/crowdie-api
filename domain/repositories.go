package domain

import (
	"gopkg.in/mgo.v2"
	"log"
	"gopkg.in/mgo.v2/bson"
	"fmt"
)

type Repository interface {
	GetGroupById(id string) (*Group, error)
}

type MongoRepository struct {
	db *mgo.Database
}

func NewRepository(mongoURL, database string) *MongoRepository {
	s, err := mgo.Dial(mongoURL)
	if err != nil {
		log.Panicf("Cannot connect to '%s': %s", mongoURL, err)
	}

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
