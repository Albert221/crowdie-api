package domain

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"github.com/google/logger"
)

type (
	Repository interface {
		CreateGroup(creator Member) (*Group, error)
		GetGroupById(id string, sec SecurityPile) (*Group, error)
		AddMemberToGroup(groupId string, member Member) (*Group, error)
		UpdateMemberRole(id string, role int8, sec SecurityPile) (*Group, error)
		UpdateMemberCoordsBit(id string, coords CoordsBit, sec SecurityPile) (*Group, error)
		KickMember(id string, sec SecurityPile) (*Group, error)
	}

	MongoRepository struct {
		db *mgo.Database
	}

	SecurityPile struct {
		Secret    string
		AndroidId string
	}
)

func NewRepository(mongoURL, database string) *MongoRepository {
	s, err := mgo.Dial(mongoURL)
	if err != nil {
		logger.Fatalf("cannot connect to '%s': %s", mongoURL, err)
	}

	s.SetMode(mgo.Monotonic, true)

	return &MongoRepository{
		db: s.DB(database),
	}
}

func (r MongoRepository) CreateGroup(creator Member) (*Group, error) {
	creator.Role = ADMIN
	group := NewGroup(creator)

	err := r.db.C("groups").Insert(group)
	if err != nil {
		return nil, fmt.Errorf("can't create new group: %s", err)
	}

	logger.Infof("created new group (%s)", group.Id)

	return &group, nil
}

func (r MongoRepository) GetGroupById(id string, sec SecurityPile) (*Group, error) {
	var group Group
	err := r.db.C("groups").Find(bson.M{"id": id}).One(&group)
	if err == mgo.ErrNotFound {
		return nil, GroupNotExists
	} else if err != nil {
		return nil, err
	}

	if !checkMemberPermissionsToGroup(&group, sec) {
		return nil, NoSufficientPermissions
	}

	return &group, nil
}

func (r MongoRepository) AddMemberToGroup(id string, member Member) (*Group, error) {
	var group Group
	query := r.db.C("groups").Find(bson.M{"id": id})
	query.One(&group)

	if r.memberWithAndroidIdExists(&group, member.AndroidId) {
		logger.Infof("an attempt to join the group (%s) has been made by the user (android id: %s)", id, member.AndroidId)

		return nil, NoSufficientPermissions
	}

	change := mgo.Change{
		Update:    bson.M{"$push": bson.M{"members": member}},
		ReturnNew: true,
	}

	_, err := query.Apply(change, &group)
	if err == mgo.ErrNotFound {
		return nil, GroupNotExists
	} else if err != nil {
		return nil, err
	}

	logger.Infof("added new member (%s) to group (%s)", member.Id, id)

	return &group, nil
}

func (r MongoRepository) UpdateMemberRole(id string, role int8, sec SecurityPile) (*Group, error) {
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"members.$.role": role}},
		ReturnNew: true,
	}

	var group Group
	query := r.db.C("groups").Find(bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}})
	query.One(&group)

	if !checkAdminPermissions(&group, sec) {
		return nil, NoSufficientPermissions
	}

	_, err := query.Apply(change, &group)
	if err == mgo.ErrNotFound {
		return nil, MemberNotExists
	} else if err != nil {
		return nil, err
	}

	logger.Infof("updated role of member (%s) to %d", id, role)

	return &group, nil
}

func (r MongoRepository) UpdateMemberCoordsBit(id string, coords CoordsBit, sec SecurityPile) (*Group, error) {
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"members.$.coordsbit": coords}},
		ReturnNew: true,
	}

	var group Group
	query := r.db.C("groups").Find(bson.M{"members": bson.M{"$elemMatch": bson.M{"id": id}}})

	err := query.One(&group)
	if err == mgo.ErrNotFound {
		return nil, MemberNotExists
	} else if err != nil {
		return nil, err
	}

	if !verifyMemberValidAndInGroup(id, &group, sec) {
		return nil, NoSufficientPermissions
	}

	_, err = query.Apply(change, &group)
	if err == mgo.ErrNotFound {
		return nil, MemberNotExists
	} else if err != nil {
		return nil, err
	}

	return &group, nil
}

// TODO: If member is last admin, give a random member an admin
func (r MongoRepository) KickMember(id string, sec SecurityPile) (*Group, error) {
	change := mgo.Change{
		Update:    bson.M{"$pull": bson.M{"members": bson.M{"id": id}}},
		ReturnNew: true,
	}

	var group Group
	query := r.db.C("groups").Find(bson.M{"members.id": id})
	query.One(&group)

	if verifyMemberValidAndInGroup(id, &group, sec) {
		// Everything's right
	} else if checkAdminPermissions(&group, sec) {
		// Everything's right too
	} else {
		return nil, NoSufficientPermissions
	}

	_, err := query.Apply(change, &group)
	if err == mgo.ErrNotFound {
		return nil, MemberNotExists
	} else if err != nil {
		return nil, err
	}

	logger.Infof("kicked member (%s) from group (%s)", id, group.Id)

	return &group, nil
}

func (r MongoRepository) memberWithAndroidIdExists(group *Group, androidId string) bool {
	for _, member := range group.Members {
		if member.AndroidId == androidId {
			return true
		}
	}

	return false
}

func checkMemberPermissionsToGroup(group *Group, sec SecurityPile) bool {
	for _, member := range group.Members {
		if member.Secret == sec.Secret && member.AndroidId == sec.AndroidId {
			return true
		}
	}

	return false
}

func verifyMemberValidAndInGroup(memberId string, group *Group, sec SecurityPile) bool {
	for _, member := range group.Members {
		if member.Id == memberId {
			if member.Secret == sec.Secret && member.AndroidId == sec.AndroidId {
				return true
			}

			return false
		}
	}

	return false
}

func checkAdminPermissions(group *Group, sec SecurityPile) bool {
	for _, member := range group.Members {
		if member.AndroidId == sec.AndroidId && member.Secret == sec.Secret {
			return member.Role == ADMIN
		}
	}

	return false
}
