package service

import "smieci-sms/internal/model"

type GarbageService interface {
    FetchDailySchedule() ([]model.GarbageSchedule, error)
}

type garbageService struct {
    sourceURL string
}

func NewGarbageService(sourceURL string) GarbageService {
    return &garbageService{sourceURL: sourceURL}
}

func (s *garbageService) FetchDailySchedule() ([]model.GarbageSchedule, error) {
    // TODO: request city garbage website, parse schedule data, return parsed results
    return []model.GarbageSchedule{}, nil
}
