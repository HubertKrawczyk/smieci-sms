package scheduler

import (
	"context"
	"log"
	"time"

	"smieci-sms/internal/model"
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
		// Run once immediately on startup so you don't wait 24h to test it
		s.runDailyJob()

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			s.runDailyJob()
		}
	}()
}

func (s *Scheduler) runDailyJob() {
	log.Println("=== Starting Daily Scheduler Job ===")

	// Create a base context with a maximum timeout of 10 minutes for the whole job
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	outdatedIDs, err := s.userRepo.GetOutdatedLocationIDs(ctx)
	if err != nil {
		log.Printf("Scheduler error: failed to fetch outdated IDs: %v", err)
		return
	}

	if len(outdatedIDs) == 0 {
		log.Println("Scheduler: All database cache schedules are up to date. No scraping needed.")
	} else {
		log.Printf("Scheduler: Found %d outdated locations. Initializing scraper sync...", len(outdatedIDs))

		freshSchedules, err := s.garbageService.FetchSchedulesForLocations(ctx, outdatedIDs)
		if err != nil {
			log.Printf("Scheduler error: scraping process failed: %v", err)
			return
		}

		if len(freshSchedules) > 0 {
			err = s.userRepo.SaveGarbageSchedules(ctx, freshSchedules)
			if err != nil {
				log.Printf("Scheduler error: failed to save fresh schedules to DB: %v", err)
				return
			}
			log.Printf("Scheduler: Successfully updated %d schedules in DB.", len(freshSchedules))
		}
	}

	log.Println("Scheduler: Checking schedules to determine tomorrow's SMS notifications...")
	s.processAndSendSMSNotifications(ctx)

	log.Println("=== Daily Scheduler Job Finished ===")
}

func (s *Scheduler) processAndSendSMSNotifications(ctx context.Context) {

}

// Helper logic to build the warning message if a date hits tomorrow
func (s *Scheduler) checkPickupTomorrow(sched *model.GarbageSchedule, tomorrow time.Time) string {
	if sched == nil {
		return ""
	}

	targetDate := tomorrow.Format("2006-01-02")
	var fractions []string

	checkDate := func(t *time.Time, name string) {
		if t != nil && t.Format("2006-01-02") == targetDate {
			fractions = append(fractions, name)
		}
	}

	checkDate(sched.DateZmieszane, "zmieszane")
	checkDate(sched.DatePapier, "papier")
	checkDate(sched.DatePlastik, "plastik i metale")
	checkDate(sched.DateSzklo, "szkło")
	checkDate(sched.DateBio, "bio")
	checkDate(sched.DateZielone, "zielone")
	checkDate(sched.DateBioRestauracyjne, "bio restauracyjne")
	checkDate(sched.DateGabaryty, "gabaryty")

	if len(fractions) == 0 {
		return ""
	}

	// Returns comma separated string if multiple pickups align on the same day
	// e.g., "bio, plastik i metale"
	return joinStrings(fractions, ", ")
}

func joinStrings(elements []string, sep string) string {
	if len(elements) == 0 {
		return ""
	}
	res := elements[0]
	for _, s := range elements[1:] {
		res += sep + s
	}
	return res
}
