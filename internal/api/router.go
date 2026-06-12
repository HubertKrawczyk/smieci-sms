package api

import (
	"github.com/gorilla/mux"
	"smieci-sms/internal/repository"
	"smieci-sms/internal/service"
)

func NewRouter(repo repository.UserRepository, garbageSvc service.GarbageService) *mux.Router {
	router := mux.NewRouter()
	handler := NewHandler(repo, garbageSvc)

	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/users", handler.CreateUserLocation).Methods("POST")
	router.HandleFunc("/users", handler.ListUsers).Methods("GET")
	router.HandleFunc("/schedules/fetch", handler.FetchSchedules).Methods("POST")

	return router
}
