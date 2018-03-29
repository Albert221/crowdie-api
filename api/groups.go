package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"wolszon.me/groupie/domain"
	"encoding/json"
)

func GetGroup(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	group, err := domain.GetGroupById(id)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(group)
}