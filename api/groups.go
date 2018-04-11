package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"wolszon.me/groupie/domain"
	"encoding/json"
	"strconv"
	"time"
	"io/ioutil"
)

var Repository domain.Repository

func NewGroup(w http.ResponseWriter, r *http.Request) {
	creator := createMemberFromRequest(r)

	group, err := Repository.NewGroup(creator)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(group)
}

func GetGroup(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	group, err := Repository.GetGroupById(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group)
}

func AddMemberToGroup(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]
	member := createMemberFromRequest(r)

	group, err := Repository.AddMemberToGroup(id, member)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	encoder.Encode(group)
}

func UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]
	role, _ := strconv.ParseInt(r.PostFormValue("role"), 10, 8)

	group, err := Repository.UpdateMemberRole(id, int8(role))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group)
}

func SendMemberCoordBit(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]
	lat, _ := strconv.ParseFloat(r.PostFormValue("lat"), 32)
	lng, _ := strconv.ParseFloat(r.PostFormValue("lng"), 32)

	coords := domain.CoordsBit{
		Lat:  float32(lat),
		Lng:  float32(lng),
		Time: time.Now(),
	}
	group, err := Repository.UpdateMemberCoordsBit(id, coords)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group)
}

func KickMember(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	group, err := Repository.KickMember(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group)
}

func createMemberFromRequest(r *http.Request) domain.Member {
	member := make(map[string]interface{})

	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &member)

	name := member["name"].(string)
	lat := float32(member["lat"].(float64))
	lng := float32(member["lng"].(float64))
	androidId := member["androidId"].(string)

	return domain.Member{
		Name: name,
		CoordsBit: domain.CoordsBit{
			Lat:  lat,
			Lng:  lng,
			Time: time.Now(),
		},
		AndroidId: androidId,
	}
}
