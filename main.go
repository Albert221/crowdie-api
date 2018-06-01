package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"wolszon.me/groupie/api"
	"fmt"
	"os"
	"github.com/google/logger"
	"wolszon.me/groupie/domain"
)

var apiController api.ApiController

func main() {
	logFile, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		logger.Errorf("error opening file: %s", err)
	}
	defer logFile.Close()

	logger.Init("Main", false, false, logFile)

	envs := getEnvs()

	apiController = api.NewApiController(
		domain.NewRepository(envs["GROUPIE_MONGO_URL"], envs["GROUPIE_DATABASE"]),
		api.NewTokenManager(envs["GROUPIE_JWT_SECRET"]),
	)

	srv := setupHttp(envs["GROUPIE_PORT"])
	fmt.Println("I'm up and running!")
	logger.Fatal(srv.ListenAndServe())
}

func getEnvs() (r map[string]string) {
	r = make(map[string]string)
	required := []string{"GROUPIE_PORT", "GROUPIE_MONGO_URL", "GROUPIE_DATABASE", "GROUPIE_JWT_SECRET"}

	for _, env := range required {
		if os.Getenv(env) == "" {
			logger.Fatalf("Environment variable %s is missing", env)
		}
		r[env] = os.Getenv(env)
	}

	return
}

func setupHttp(port string) *http.Server {
	r := mux.NewRouter()

	apiRoutes := r.PathPrefix("/api/v1").Subrouter()
	apiRoutes.Use(apiController.GetMiddleware())

	apiRoutes.HandleFunc("/group", apiController.NewGroup).
		Methods("POST").
		Name("group.create")

	apiRoutes.HandleFunc("/group/{id}", apiController.JoinGroup).
		Methods("POST").
		Name("group.join")

	apiRoutes.HandleFunc("/group/{id}", apiController.Endpoint).
		Methods("GET").
		Name("group.endpoint")

	return &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}
