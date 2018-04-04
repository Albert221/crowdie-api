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
	envs := getEnvs()

	api.Repository = domain.NewRepository(
		envs["GROUPIE_MONGO_URL"], envs["GROUPIE_DATABASE"])

	srv := setupHttp(envs["GROUPIE_PORT"])
	log.Fatal(srv.ListenAndServe())
}

func getEnvs() (r map[string]string) {
	r = make(map[string]string)
	required := []string{"GROUPIE_PORT", "GROUPIE_MONGO_URL", "GROUPIE_DATABASE"}

	for _, env := range required {
		if os.Getenv(env) == "" {
			log.Panicf("Environment variable %s is missing", env)
		}
		r[env] = os.Getenv(env)
	}

	return
}

func setupHttp(port string) *http.Server {
	r := mux.NewRouter()

	r.HandleFunc("/group", api.NewGroup).
		Methods("POST").
		Name("group.create")

	r.HandleFunc("/group/{id}", api.GetGroup).
		Methods("GET").
		Name("group.get")

	r.HandleFunc("/group/{id}/member", api.AddMemberToGroup).
		Methods("POST").
		Name("group.addMember")

	r.HandleFunc("/member/{id}/role", api.UpdateMemberRole).
		Methods("PATCH").
		Name("member.updateRole")

	r.HandleFunc("/member/{id}/coords-bit", api.SendMemberCoordBit).
		Methods("PATCH").
		Name("member.coordsBit")

	r.HandleFunc("/member/{id}", api.KickMember).
		Methods("DELETE").
		Name("member.kick")

	return &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}
