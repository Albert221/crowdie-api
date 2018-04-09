package domain

import (
	"gopkg.in/mgo.v2"
	"log"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"github.com/google/uuid"
)

type Repository interface {
	NewGroup(creator Member) (*Group, error)
	GetGroupById(id string) (*Group, error)
	AddMemberToGroup(groupId string, member Member) (*Group, error)
	UpdateMemberRole(id string, role int8) (*Member, error)
	UpdateMemberCoordsBit(id string, coords CoordsBit) (*Member, error)
	KickMember(id string) (*Group, error)
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

func (r MongoRepository) NewGroup(creator Member) (*Group, error) {
	creator.Id = uuid.New().String()
	creator.Role = ADMIN

	group := Group{
		Id: uuid.New().String(),
		Members: []Member{
			creator,
		},
	}

	err := r.db.C("groups").Insert(group)
	if err != nil {
		return nil, fmt.Errorf("can't insert new group: %s", err)
	}

	return &group, nil
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

func (r MongoRepository) AddMemberToGroup(id string, member Member) (*Group, error) {
	member.Id = uuid.New().String()

	change := mgo.Change{
		Update:    bson.M{"$push": bson.M{"members": member}},
		ReturnNew: true,
	}

	var updatedGroup Group
	_, err := r.db.C("groups").Find(bson.M{"id": id}).Apply(change, &updatedGroup)
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("group with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	return &updatedGroup, nil
}

func (r MongoRepository) UpdateMemberRole(id string, role int8) (*Member, error) {
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"members.$.role": role}},
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

func (r MongoRepository) UpdateMemberCoordsBit(id string, coords CoordsBit) (*Member, error) {
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"members.$.coordsbit": coords}},
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

func (r MongoRepository) KickMember(id string) (*Group, error) {
	change := mgo.Change{
		Update:    bson.M{"$pull": bson.M{"members": bson.M{"id": id}}},
		ReturnNew: true,
	}

	var updatedGroup Group
	_, err := r.db.C("groups").Find(bson.M{"members.id": id}).Apply(change, &updatedGroup)
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("member with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	return &updatedGroup, nil
}
