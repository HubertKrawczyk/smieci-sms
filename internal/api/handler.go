package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"smieci-sms/internal/model"
	"smieci-sms/internal/repository"
	"smieci-sms/internal/service"
)

type Handler struct {
	repo       repository.UserRepository
	garbageSvc service.GarbageService
}

func NewHandler(repo repository.UserRepository, garbageSvc service.GarbageService) *Handler {
	return &Handler{repo: repo, garbageSvc: garbageSvc}
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

	user := model.UserLocation{
		Name:        payload.Name,
		Phone:       payload.Phone,
		LocationID:  payload.LocationID,
		AddressName: payload.AddressName,
	}
	fmt.Printf("User struct: %+v\n", user)

	if err := h.repo.SaveUserLocation(r.Context(), user); err != nil {
		log.Printf("ERROR: SaveUserLocation failed: %v", err)
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

func (h *Handler) FetchSchedules(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		LocationIDs      []int `json:"location_ids"`
		FetchAllExisting bool  `json:"fetch_all_existing"`
	}

	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	// Decouple context from client disconnects/timeouts
	ctx := context.WithoutCancel(r.Context())

	var locationIDs []int
	if payload.FetchAllExisting {
		users, err := h.repo.ListUsers(ctx)
		if err != nil {
			log.Printf("ERROR: ListUsers failed: %v", err)
			http.Error(w, "failed to list users", http.StatusInternalServerError)
			return
		}
		// Collect unique location IDs
		seen := make(map[int]bool)
		for _, u := range users {
			locID := int(u.LocationID)
			if !seen[locID] {
				seen[locID] = true
				locationIDs = append(locationIDs, locID)
			}
		}
	} else if len(payload.LocationIDs) > 0 {
		locationIDs = payload.LocationIDs
	} else {
		// Fallback: get outdated location IDs
		ids, err := h.repo.GetOutdatedLocationIDs(ctx)
		if err != nil {
			log.Printf("ERROR: GetOutdatedLocationIDs failed: %v", err)
			http.Error(w, "failed to get outdated location IDs", http.StatusInternalServerError)
			return
		}
		locationIDs = ids
	}

	if len(locationIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]model.GarbageSchedule{})
		return
	}

	schedules, err := h.garbageSvc.FetchSchedulesForLocations(ctx, locationIDs)
	if err != nil {
		log.Printf("ERROR: FetchSchedulesForLocations failed: %v", err)
		http.Error(w, fmt.Sprintf("failed to fetch schedules: %v", err), http.StatusInternalServerError)
		return
	}

	if len(schedules) > 0 {
		if err := h.repo.SaveGarbageSchedules(ctx, schedules); err != nil {
			log.Printf("ERROR: SaveGarbageSchedules failed: %v", err)
			http.Error(w, "failed to save schedules to database", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}
