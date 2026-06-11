package api

import (
    "encoding/json"
    "net/http"

    "smieci-sms/internal/model"
    "smieci-sms/internal/repository"
)

type Handler struct {
    repo repository.UserRepository
}

func NewHandler(repo repository.UserRepository) *Handler {
    return &Handler{repo: repo}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) CreateUserLocation(w http.ResponseWriter, r *http.Request) {
    var payload model.UserLocationRequest
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    user := model.User{
        Name:  payload.Name,
        Phone: payload.Phone,
    }

    if err := h.repo.SaveUser(r.Context(), user, payload.Address); err != nil {
        http.Error(w, "failed to save user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.repo.ListUsers(r.Context())
    if err != nil {
        http.Error(w, "failed to list users", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}
