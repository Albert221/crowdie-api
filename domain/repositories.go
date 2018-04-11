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
	UpdateMemberRole(id string, role int8) (*Group, error)
	UpdateMemberCoordsBit(id string, coords CoordsBit) (*Group, error)
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
	var group Group
	query := r.db.C("groups").Find(bson.M{"id": id})
	query.One(&group)

	exists, id := r.memberWithAndroidIdExists(&group, member.AndroidId)

	var err error
	if exists {
		return r.UpdateMemberCoordsBit(id, member.CoordsBit)
	}

	member.Id = uuid.New().String()
	change := mgo.Change{
		Update:    bson.M{"$push": bson.M{"members": member}},
		ReturnNew: true,
	}
	_, err = query.Apply(change, &group)

	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("group with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

func (r MongoRepository) memberWithAndroidIdExists(group *Group, androidId string) (bool, string) {
	for _, member := range group.Members {
		if member.AndroidId == androidId {
			return true, member.Id
		}
	}

	return false, ""
}

func (r MongoRepository) UpdateMemberRole(id string, role int8) (*Group, error) {
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"members.$.role": role}},
		ReturnNew: true,
	}

	var group Group
	_, err := r.db.C("groups").Find(bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}}).Apply(change, &group)
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("member with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

func (r MongoRepository) UpdateMemberCoordsBit(id string, coords CoordsBit) (*Group, error) {
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"members.$.coordsbit": coords}},
		ReturnNew: true,
	}

	var group Group
	_, err := r.db.C("groups").Find(bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}}).Apply(change, &group)
	if err == mgo.ErrNotFound {
		return nil, fmt.Errorf("member with ID '%s' does not exist", id)
	} else if err != nil {
		return nil, err
	}

	return &group, nil
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
