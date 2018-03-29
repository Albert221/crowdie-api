package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"time"
	"wolszon.me/groupie/api"
	"wolszon.me/groupie/domain"
	"fmt"
	"os"
)

func main() {
	domain.SetupMockData()

	r := mux.NewRouter()
	r.HandleFunc("/group/{id}", api.GetGroup).
		Methods("GET").
		Name("group.get")

	var port string
	if port = os.Getenv("GROUPIE_PORT"); port == "" {
		port = "8080"
	}

	srv := http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
