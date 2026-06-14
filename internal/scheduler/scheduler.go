package scheduler

import (
	"context"
	"fmt"
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
	telegramSvc    service.TelegramService
}

func NewScheduler(userRepo repository.UserRepository, garbageService service.GarbageService, smsService service.SMSService, telegramSvc service.TelegramService) *Scheduler {
	return &Scheduler{
		userRepo:       userRepo,
		garbageService: garbageService,
		smsService:     smsService,
		telegramSvc:    telegramSvc,
	}
}

func (s *Scheduler) ScheduleDailyTasks() {
	// 1. Daily scraper sync loop
	go func() {
		// Run once immediately on startup so you don't wait 24h to test it
		s.runDailyJob()

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			s.runDailyJob()
		}
	}()

	// 2. Hourly notifications check loop
	go func() {
		// Run once immediately on startup so notifications check runs at start
		s.runHourlyJob()

		// Align to the top of the next hour
		now := time.Now()
		nextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
		time.Sleep(nextHour.Sub(now))

		s.runHourlyJob() // run exactly at top of the hour

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			s.runHourlyJob()
		}
	}()
}

func (s *Scheduler) runDailyJob() {
	log.Println("=== Starting Daily Scraper Scheduler Job ===")

	// Create a base context with a maximum timeout of 10 minutes for the whole job
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	outdatedIDs, err := s.userRepo.GetOutdatedLocationIDs(ctx)
	if err != nil {
		log.Printf("Scheduler error: failed to fetch outdated IDs: %v", err)
		return
	}

	if len(outdatedIDs) == 0 {
		log.Println("Scheduler: All database cache schedules are up to date.")
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

	log.Println("Scheduler: Cleaning up orphaned garbage schedules...")
	if err := s.userRepo.DeleteOrphanedSchedules(ctx); err != nil {
		log.Printf("Scheduler error: failed to clean up orphaned schedules: %v", err)
	}

	log.Println("=== Daily Scraper Scheduler Job Finished ===")
}

func (s *Scheduler) runHourlyJob() {
	log.Println("=== Starting Hourly Scheduler Job ===")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	now := time.Now()
	hour := now.Hour()

	// 1. Process notifications for today (morning_X)
	s.processNotifications(ctx, fmt.Sprintf("morning_%d", hour), false, now)

	// 2. Process notifications for tomorrow (day_before_X)
	s.processNotifications(ctx, fmt.Sprintf("day_before_%d", hour), true, now)

	log.Println("=== Hourly Scheduler Job Finished ===")
}

func (s *Scheduler) processNotifications(ctx context.Context, targetPref string, isTomorrow bool, now time.Time) {
	log.Printf("Scheduler: Processing notifications for preference %q...", targetPref)

	userSchedules, err := s.userRepo.ListUsersWithPreferenceAndSchedule(ctx, targetPref)
	if err != nil {
		log.Printf("Scheduler error: failed to fetch users with preference %q: %v", targetPref, err)
		return
	}

	if len(userSchedules) == 0 {
		return
	}

	log.Printf("Scheduler: Found %d users with preference %q.", len(userSchedules), targetPref)

	var targetDate time.Time
	if isTomorrow {
		targetDate = now.AddDate(0, 0, 1)
	} else {
		targetDate = now
	}

	for _, us := range userSchedules {
		fractions := s.checkPickupForDate(&us.Schedule, targetDate)
		if len(fractions) == 0 {
			continue
		}

		fractionsStr := joinStrings(fractions, ", ")
		var msg string
		if isTomorrow {
			msg = fmt.Sprintf("Przypomnienie: jutro (%s) odbiór następujących odpadów: %s. Pamiętaj o wystawieniu koszy/worków!", targetDate.Format("02.01.2006"), fractionsStr)
		} else {
			msg = fmt.Sprintf("Dzień dobry! Dzisiaj (%s) odbiór następujących odpadów: %s. Upewnij się, że kosze/worki są wystawione!", targetDate.Format("02.01.2006"), fractionsStr)
		}

		if us.User.ChatID != -1 {
			log.Printf("Sending Telegram notification to user %s (ChatID: %d)", us.User.Name, us.User.ChatID)
			s.sendTelegramMessage(us.User.ChatID, msg)
		}

		if s.smsService != nil && us.User.Phone != "" && us.User.Phone != "123456789" {
			log.Printf("Sending SMS notification to user %s (Phone: %s)", us.User.Name, us.User.Phone)
			if err := s.smsService.SendSMS(us.User.Phone, msg); err != nil {
				log.Printf("Failed to send SMS to %s: %v", us.User.Phone, err)
			}
		}
	}
}

func (s *Scheduler) checkPickupForDate(sched *model.GarbageSchedule, targetDate time.Time) []string {
	if sched == nil {
		return nil
	}

	dateStr := targetDate.Format("2006-01-02")
	var fractions []string

	checkDate := func(t *time.Time, name string) {
		if t != nil && t.Format("2006-01-02") == dateStr {
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

	return fractions
}

func (s *Scheduler) sendTelegramMessage(chatID int64, text string) {
	if err := s.telegramSvc.SendMessage(context.Background(), chatID, text, nil); err != nil {
		log.Printf("Scheduler error sending Telegram message: %v", err)
	}
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
