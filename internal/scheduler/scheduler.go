package scheduler

import (
    "time"

    "smieci-sms/internal/repository"
    "smieci-sms/internal/service"
)

type Scheduler struct {
    userRepo       repository.UserRepository
    garbageService service.GarbageService
    smsService     service.SMSService
}

func NewScheduler(userRepo repository.UserRepository, garbageService service.GarbageService, smsService service.SMSService) *Scheduler {
    return &Scheduler{userRepo: userRepo, garbageService: garbageService, smsService: smsService}
}

func (s *Scheduler) ScheduleDailyTasks() {
    go func() {
        ticker := time.NewTicker(24 * time.Hour)
        defer ticker.Stop()

        for {
            s.runDailyJob()
            <-ticker.C
        }
    }()
}

func (s *Scheduler) runDailyJob() {
    // TODO: fetch schedule from city website, compare with saved users, send SMS when collection occurs
}
