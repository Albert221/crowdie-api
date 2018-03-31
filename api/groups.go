package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"wolszon.me/groupie/domain"
	"encoding/json"
	"strconv"
	"time"
)

var Repository domain.Repository

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

func UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]
	role, _ := strconv.ParseInt(r.PostFormValue("role"), 10, 8)

	member, err := Repository.UpdateMemberRole(id, int8(role))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(member)
}

func SendMemberCoordBit(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]
	lat, _ := strconv.ParseFloat(r.PostFormValue("lat"), 32)
	lng, _ := strconv.ParseFloat(r.PostFormValue("lng"), 32)

	member, err := Repository.UpdateMemberCoordsBit(id, float32(lat), float32(lng), time.Now())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(member)
}