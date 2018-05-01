package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"wolszon.me/groupie/api"
	"wolszon.me/groupie/domain"
	"fmt"
	"os"
	"github.com/google/logger"
)

func main() {
	logFile, _ := os.OpenFile("log.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0664)
	defer logFile.Close()

	logger.Init("Main", false, false, logFile)

	envs := getEnvs()

	api.Repository = domain.NewRepository(envs["GROUPIE_MONGO_URL"], envs["GROUPIE_DATABASE"])

	srv := setupHttp(envs["GROUPIE_PORT"])
	logger.Fatal(srv.ListenAndServe())
}

func getEnvs() (r map[string]string) {
	r = make(map[string]string)
	required := []string{"GROUPIE_PORT", "GROUPIE_MONGO_URL", "GROUPIE_DATABASE"}

	for _, env := range required {
		if os.Getenv(env) == "" {
			logger.Errorf("Environment variable %s is missing", env)
		}
		r[env] = os.Getenv(env)
	}

	return
}

func setupHttp(port string) *http.Server {
	r := mux.NewRouter()

	apiRoutes := r.PathPrefix("/api/v1").Subrouter()

	apiRoutes.HandleFunc("/group", api.NewGroup).
		Methods("POST").
		Name("group.create")

	apiRoutes.HandleFunc("/group/{id}", api.GetGroup).
		Methods("GET").
		Name("group.get")

	apiRoutes.HandleFunc("/group/{id}/member", api.AddMemberToGroup).
		Methods("POST").
		Name("group.addMember")

	apiRoutes.HandleFunc("/member/{id}/role", api.UpdateMemberRole).
		Methods("PATCH").
		Name("member.updateRole")

	apiRoutes.HandleFunc("/member/{id}/coords-bit", api.SendMemberCoordBit).
		Methods("PATCH").
		Name("member.coordsBit")

	apiRoutes.HandleFunc("/member/{id}", api.KickMember).
		Methods("DELETE").
		Name("member.kick")

	return &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}
