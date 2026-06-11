package api

import (
    "github.com/gorilla/mux"
    "smieci-sms/internal/repository"
)

func NewRouter(repo repository.UserRepository) *mux.Router {
    router := mux.NewRouter()
    handler := NewHandler(repo)

    router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
    router.HandleFunc("/users", handler.CreateUserLocation).Methods("POST")
    router.HandleFunc("/users", handler.ListUsers).Methods("GET")

    return router
}
