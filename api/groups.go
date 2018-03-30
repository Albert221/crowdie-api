package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"wolszon.me/groupie/domain"
	"encoding/json"
)

var Repository domain.Repository

func GetGroup(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	group, err := Repository.GetGroupById(id)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group)
}
