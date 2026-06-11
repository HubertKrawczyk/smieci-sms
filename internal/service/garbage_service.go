package service

import (
	"context"
	"log"
	"net/http"
	"smieci-sms/internal/model"
	"time"
)

type GarbageService interface {
	FetchSchedulesForLocations(ctx context.Context, locationIDs []int) ([]model.GarbageSchedule, error)
}

type garbageService struct {
	sourceURL  string
	httpClient *http.Client
}

func NewGarbageService(sourceURL string) GarbageService {
	return &garbageService{
		sourceURL:  sourceURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *garbageService) FetchSchedulesForLocations(ctx context.Context, locationIDs []int) ([]model.GarbageSchedule, error) {
	var updatedSchedules []model.GarbageSchedule

	for _, locID := range locationIDs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		log.Printf("Refreshing schedule from Warsaw website for location: %d", locID)

		// 1. TODO: Make HTTP request to s.sourceURL using locID
		// 2. TODO: Parse the updated HTML/JSON dates from the city response

		freshData := model.GarbageSchedule{
			LocationID: locID,
			// DateZmieszane: parsedDate, // map your parsed fields here
			LastUpdate: time.Now(),
		}

		updatedSchedules = append(updatedSchedules, freshData)

		// DevOps courtesy rule: wait 500ms between requests so the city doesn't block your IP
		time.Sleep(500 * time.Millisecond)
	}

	return updatedSchedules, nil
}
