package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"wolszon.me/groupie/domain"
	"encoding/json"
	"strconv"
	"time"
	"io/ioutil"
	"github.com/google/logger"
	"github.com/gbrlsnchs/jwt"
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

var (
	Repository      domain.Repository
	ApiTokenManager TokenManager
)

func NewGroup(w http.ResponseWriter, r *http.Request) {
	creator := createMemberFromRequest(r)

	group, err := Repository.CreateGroup(creator)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := ApiTokenManager.CreateToken(creator.Secret, creator.AndroidId)
	if err != nil {
		logger.Warning(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := bson.M{
		"yourId": creator.Id,
		"group": group.Export(),
		"token": token,
	}

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

func GetGroup(w http.ResponseWriter, r *http.Request) {
	secPile, err := getSecurityPileFromJWT(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	v := mux.Vars(r)
	id := v["id"]

	group, err := Repository.GetGroupById(id, secPile)
	if err == domain.GroupNotExists {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err == domain.NoSufficientPermissions {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group.Export())
}

func AddMemberToGroup(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]
	member := createMemberFromRequest(r)

	existsFlag := new(bool)
	group, err := Repository.AddMemberToGroup(id, member, func() (domain.SecurityPile, error) {
		secPile, err := getSecurityPileFromJWT(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return secPile, err
		}

		return secPile, nil
	}, existsFlag)
	if err == domain.GroupNotExists {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err == domain.NoSufficientPermissions {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := bson.M{
		"group": group.Export(),
	}

	if !*existsFlag {
		token, err := ApiTokenManager.CreateToken(member.Secret, member.AndroidId)
		if err != nil {
			logger.Warning(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response["yourId"] = member.Id
		response["token"] = token
	}

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

func UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	secPile, err := getSecurityPileFromJWT(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	v := mux.Vars(r)
	id := v["id"]
	role, _ := strconv.ParseInt(r.PostFormValue("role"), 10, 8)

	group, err := Repository.UpdateMemberRole(id, int8(role), secPile)
	if err == domain.MemberNotExists {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err == domain.NoSufficientPermissions {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group.Export())
}

func SendMemberCoordBit(w http.ResponseWriter, r *http.Request) {
	secPile, err := getSecurityPileFromJWT(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	v := mux.Vars(r)
	id := v["id"]
	lat, _ := strconv.ParseFloat(r.PostFormValue("lat"), 32)
	lng, _ := strconv.ParseFloat(r.PostFormValue("lng"), 32)

	coords := domain.CoordsBit{
		Lat:  float32(lat),
		Lng:  float32(lng),
		Time: time.Now(),
	}
	group, err := Repository.UpdateMemberCoordsBit(id, coords, secPile)
	if err == domain.MemberNotExists {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err == domain.NoSufficientPermissions {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group.Export())
}

func KickMember(w http.ResponseWriter, r *http.Request) {
	secPile, err := getSecurityPileFromJWT(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	v := mux.Vars(r)
	id := v["id"]

	group, err := Repository.KickMember(id, secPile)
	if err == domain.MemberNotExists {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err == domain.NoSufficientPermissions {
		w.WriteHeader(http.StatusForbidden)
		return
	} else if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group.Export())
}

func createMemberFromRequest(r *http.Request) domain.Member {
	member := make(map[string]interface{})

	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &member)

	name := member["name"].(string)
	androidId := member["androidId"].(string)
	lat := float32(member["lat"].(float64))
	lng := float32(member["lng"].(float64))

	return domain.NewMember(name, androidId, lat, lng)
}

func getSecurityPileFromJWT(r *http.Request) (domain.SecurityPile, error) {
	t := r.Context().Value(ContextJwt)
	if t == nil {
		return domain.SecurityPile{}, fmt.Errorf("can't retrieve JWT from request")
	}

	token := t.(*jwt.JWT)
	public := token.Public()

	secret, secretExists := public["secret"]
	androidId, androidIdExists := public["androidId"]
	if !secretExists || !androidIdExists {
		return domain.SecurityPile{}, fmt.Errorf("JWT payload lacks secret and/or androidId")
	}

	return domain.SecurityPile{
		Secret:    secret.(string),
		AndroidId: androidId.(string),
	}, nil
}
