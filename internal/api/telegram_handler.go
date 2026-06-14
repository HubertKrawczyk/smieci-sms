package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"smieci-sms/internal/commands"
	"smieci-sms/internal/messages"
	"smieci-sms/internal/model"
	"smieci-sms/internal/repository"
	"smieci-sms/internal/service"
)

type TelegramHandler struct {
	repo              repository.UserRepository
	garbageService    service.GarbageService
	secretToken       string
	telegramSvc       service.TelegramService
	registrationQueue chan string
}

type ConversationState int

const (
	StateNone ConversationState = iota
	StateAwaitingStreet
	StateAwaitingNumber
	StateAwaitingPostcode
	StateAwaitingLocationConfirmation
	StateAwaitingSchedule
)

type UserSession struct {
	State               ConversationState
	Street              string
	Number              string
	Postcode            string
	LocationID          string
	LocationName        string
	SelectedPreferences []string
}

var (
	sessionMutex sync.Mutex // concurrency safety
	sessions     = make(map[int64]*UserSession)
)

func getOrCreateSession(chatID int64) *UserSession {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	if _, exists := sessions[chatID]; !exists { // create new session if not exists
		sessions[chatID] = &UserSession{State: StateNone}
	}
	return sessions[chatID]
}

func NewTelegramHandler(repo repository.UserRepository, garbageService service.GarbageService, secretToken string, telegramSvc service.TelegramService) *TelegramHandler {
	h := &TelegramHandler{
		repo:              repo,
		garbageService:    garbageService,
		secretToken:       secretToken,
		telegramSvc:       telegramSvc,
		registrationQueue: make(chan string, 200), // Buffer for up to 200 concurrent registrations
	}
	go h.processRegistrationQueue()
	return h
}

func (h *TelegramHandler) processRegistrationQueue() {
	for locID := range h.registrationQueue {
		log.Printf("Processing initial schedule fetch for location %s", locID)
		schedules, err := h.garbageService.FetchSchedulesForLocations(context.Background(), []string{locID})
		if err != nil {
			log.Printf("ERROR: failed to fetch schedule for new location %s: %v", locID, err)
			continue
		}
		if len(schedules) > 0 {
			if err := h.repo.SaveGarbageSchedules(context.Background(), schedules); err != nil {
				log.Printf("ERROR: failed to save fetched schedule to DB for location %s: %v", locID, err)
			} else {
				log.Printf("Successfully fetched and saved schedule for new location %s", locID)
			}
		}
	}
}

func (h *TelegramHandler) Start(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	receivedToken := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	log.Printf("New webhook request for start.")

	if subtle.ConstantTimeCompare([]byte(receivedToken), []byte(h.secretToken)) != 1 {
		log.Printf("Unauthorized webhook request blocked: invalid or missing secret token.")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload model.TelegramRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if payload.CallbackQuery != nil {
		cb := payload.CallbackQuery
		chatID := cb.Message.Chat.ID
		session := getOrCreateSession(chatID)

		if cb.Data == "delete_confirm" {
			h.sendCallbackAcknowledgment(cb.ID)
			if err := h.repo.DeleteUserLocationByChatID(r.Context(), chatID); err != nil {
				log.Printf("ERROR: failed to delete user location for chat ID %d: %v", chatID, err)
			}
			sessionMutex.Lock()
			delete(sessions, chatID)
			sessionMutex.Unlock()

			h.sendTelegramEditMessage(chatID, cb.Message.MessageID, messages.DataDeleted)
			return
		}

		if cb.Data == "delete_cancel" {
			h.sendCallbackAcknowledgment(cb.ID)
			h.sendTelegramEditMessage(chatID, cb.Message.MessageID, messages.DeleteCanceledMessage)
			return
		}

		if session.State == StateAwaitingLocationConfirmation {
			// Acknowledge click immediately to clear loading icon on user screen
			h.sendCallbackAcknowledgment(cb.ID)

			if cb.Data == "loc_cancel" {
				session.State = StateNone
				h.sendTelegramMessage(chatID, messages.RegistrationCanceled)
				return
			}

			if !strings.HasPrefix(cb.Data, "loc_") {
				return // Ignore clicks on older buttons (e.g., from schedule menu)
			}

			// Extract your AddressPointID from the payload string
			// Format was: "loc_YOUR_ID"
			selectedLocationID := strings.TrimPrefix(cb.Data, "loc_")
			selectedLocationName := "Selected Location"
			if cb.Message.ReplyMarkup != nil {
				for _, row := range cb.Message.ReplyMarkup.InlineKeyboard {
					for _, btn := range row {
						if btn.CallbackData == cb.Data {
							selectedLocationName = btn.Text
							break
						}
					}
				}
			}

			// Store selected location in user session
			session.LocationID = selectedLocationID
			session.LocationName = selectedLocationName
			session.SelectedPreferences = []string{} // Reset selected preferences

			// Change state to schedule options
			session.State = StateAwaitingSchedule

			// Edit original button grid text away into a plain confirmed message string
			h.sendTelegramEditMessage(chatID, cb.Message.MessageID, fmt.Sprintf(messages.SelectedLocationEdit, selectedLocationName))

			// Send new message with schedule choices
			keyboard := h.buildScheduleKeyboard(session.SelectedPreferences)
			h.sendTelegramMessage(chatID, messages.SchedulePrompt, keyboard)
			return
		}

		if session.State == StateAwaitingSchedule {
			h.sendCallbackAcknowledgment(cb.ID)

			if strings.HasPrefix(cb.Data, "pref_") {
				pref := strings.TrimPrefix(cb.Data, "pref_")

				if pref == "done" {
					// Save the registration to DB!
					// 1. Delete previous registration if any
					if err := h.repo.DeleteUserLocationByChatID(r.Context(), chatID); err != nil {
						log.Printf("WARNING: failed to delete existing user location: %v", err)
					}

					// 2. Insert new registration with SelectedPreferences
					err := h.repo.SaveUserLocation(r.Context(), model.UserLocation{
						ChatID:               chatID,
						LocationID:           session.LocationID,
						Name:                 "Użytkownik",
						Phone:                "123456789",
						AddressName:          session.LocationName,
						NotificationSettings: session.SelectedPreferences,
					})

					if err != nil {
						log.Printf("ERROR: failed to save user location: %v", err)
						h.sendTelegramEditMessage(chatID, cb.Message.MessageID, messages.SaveError)
					} else {
						// Send to queue non-blockingly so we fetch schedule in background
						select {
						case h.registrationQueue <- session.LocationID:
						default:
							log.Printf("WARNING: registration queue full, skipping immediate fetch for %s", session.LocationID)
						}

						// Edit the original message to remove buttons and show confirmation
						var prefsStr []string
						for _, p := range session.SelectedPreferences {
							prefsStr = append(prefsStr, formatPreference(p))
						}
						var finalMsg string
						if len(prefsStr) > 0 {
							finalMsg = fmt.Sprintf(messages.ConfirmationWithPreferences, session.LocationName, strings.Join(prefsStr, ", "))
						} else {
							finalMsg = fmt.Sprintf(messages.ConfirmationNoPreferences, session.LocationName)
						}
						h.sendTelegramEditMessage(chatID, cb.Message.MessageID, finalMsg)
					}

					// Reset session state
					session.State = StateNone
					session.SelectedPreferences = nil
					return
				}

				// Toggle preference
				found := false
				idx := -1
				for i, p := range session.SelectedPreferences {
					if p == pref {
						found = true
						idx = i
						break
					}
				}

				if found {
					// Remove
					session.SelectedPreferences = append(session.SelectedPreferences[:idx], session.SelectedPreferences[idx+1:]...)
				} else {
					// Add
					session.SelectedPreferences = append(session.SelectedPreferences, pref)
				}

				// Rebuild and edit reply markup to show checkmarks updated!
				newMenu := h.buildScheduleKeyboard(session.SelectedPreferences)
				h.sendTelegramEditReplyMarkup(chatID, cb.Message.MessageID, newMenu)
			}
			return
		}
	}

	if payload.Message != nil && payload.Message.Text == commands.Usun {
		chatID := payload.Message.Chat.ID
		log.Printf("User on Chat ID %d wants to delete their data via %s", chatID, commands.Usun)

		keyboard := &model.TelegramInlineMenu{
			InlineKeyboard: [][]model.TelegramInlineButton{
				{
					{Text: messages.DeleteConfirmButton, CallbackData: "delete_confirm"},
				},
				{
					{Text: messages.DeleteCancelButton, CallbackData: "delete_cancel"},
				},
			},
		}

		h.sendTelegramMessage(chatID, messages.DeleteConfirmationPrompt, keyboard)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	if payload.Message != nil && payload.Message.Text == commands.Harmonogram {
		chatID := payload.Message.Chat.ID
		log.Printf("User on Chat ID %d requested their schedule via %s", chatID, commands.Harmonogram)

		schedule, err := h.repo.GetUserScheduleByChatID(r.Context(), chatID)
		if err != nil {
			log.Printf("ERROR: failed to get schedule for chat ID %d: %v", chatID, err)
			h.sendTelegramMessage(chatID, messages.HarmonogramError)
		} else if schedule == nil {
			h.sendTelegramMessage(chatID, messages.HarmonogramNotRegistered)
		} else if schedule.Schedule.LocationID == "" {
			h.sendTelegramMessage(chatID, messages.HarmonogramPending)
		} else {
			var sb strings.Builder

			if schedule.User.AddressName != "" {
				sb.WriteString(fmt.Sprintf("📍 Adres: %s\n", schedule.User.AddressName))
			}

			if len(schedule.User.NotificationSettings) > 0 {
				var prefsStr []string
				for _, p := range schedule.User.NotificationSettings {
					prefsStr = append(prefsStr, formatPreference(p))
				}
				sb.WriteString(fmt.Sprintf("🔔 Powiadomienia: %s\n\n", strings.Join(prefsStr, ", ")))
			} else {
				sb.WriteString("🔔 Powiadomienia: Brak\n\n")
			}

			if !schedule.Schedule.LastUpdate.IsZero() {
				sb.WriteString(fmt.Sprintf(messages.HarmonogramHeaderDate, schedule.Schedule.LastUpdate.Format("2006-01-02 15:04")))
			} else {
				sb.WriteString(messages.HarmonogramHeaderNoDate)
			}

			type ScheduleEntry struct {
				Name string
				Icon string
				Date *time.Time
			}

			entries := []ScheduleEntry{
				{messages.FractionZmieszane, "⚫", schedule.Schedule.DateZmieszane},
				{messages.FractionPapier, "🔵", schedule.Schedule.DatePapier},
				{messages.FractionPlastik, "🟡", schedule.Schedule.DatePlastik},
				{messages.FractionSzklo, "🟢", schedule.Schedule.DateSzklo},
				{messages.FractionBio, "🟤", schedule.Schedule.DateBio},
				{messages.FractionZielone, "🌿", schedule.Schedule.DateZielone},
				{messages.FractionBioRestauracyjne, "🥗", schedule.Schedule.DateBioRestauracyjne},
				{messages.FractionGabaryty, "🛋️", schedule.Schedule.DateGabaryty},
			}

			sort.Slice(entries, func(i, j int) bool {
				if entries[i].Date == nil && entries[j].Date == nil {
					return entries[i].Name < entries[j].Name
				}
				if entries[i].Date == nil {
					return false // nil goes to the end
				}
				if entries[j].Date == nil {
					return true
				}
				return entries[i].Date.Before(*entries[j].Date)
			})

			for _, e := range entries {
				dateStr := messages.HarmonogramNoData
				if e.Date != nil {
					dateStr = e.Date.Format("2006-01-02") + "  "
				}
				sb.WriteString(fmt.Sprintf("%s %s %s\n", dateStr, e.Icon, e.Name))
			}

			h.sendTelegramMessage(chatID, strings.TrimSpace(sb.String()))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	if payload.Message != nil && payload.Message.Text == commands.EdytujPowiadomienia {
		chatID := payload.Message.Chat.ID
		log.Printf("User on Chat ID %d wants to edit schedule via %s", chatID, commands.EdytujPowiadomienia)

		schedule, err := h.repo.GetUserScheduleByChatID(r.Context(), chatID)
		if err != nil {
			h.sendTelegramMessage(chatID, messages.HarmonogramError)
		} else if schedule == nil {
			h.sendTelegramMessage(chatID, messages.HarmonogramNotRegistered)
		} else {
			session := getOrCreateSession(chatID)
			session.State = StateAwaitingSchedule
			session.LocationID = schedule.User.LocationID
			session.LocationName = schedule.User.AddressName
			// Make a copy of preferences to avoid accidental shared slice modification
			session.SelectedPreferences = append([]string{}, schedule.User.NotificationSettings...)

			keyboard := h.buildScheduleKeyboard(session.SelectedPreferences)
			h.sendTelegramMessage(chatID, messages.SchedulePrompt, keyboard)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	if payload.Message != nil && (payload.Message.Text == commands.Prywatnosc || payload.Message.Text == commands.Privacy) {
		chatID := payload.Message.Chat.ID
		log.Printf("User on Chat ID %d requested privacy policy", chatID)

		h.sendTelegramMessage(chatID, messages.PrivacyPolicy)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	if payload.Message != nil && payload.Message.Text == commands.Start {
		chatID := payload.Message.Chat.ID
		fmt.Printf("User on Chat ID %d wants to START the process!\n", chatID)

		session := getOrCreateSession(chatID)
		session.State = StateAwaitingStreet
		h.sendTelegramMessage(chatID, messages.Welcome)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}
	if payload.Message != nil {
		chatID := payload.Message.Chat.ID
		session := getOrCreateSession(chatID)
		text := payload.Message.Text

		switch session.State {
		case StateAwaitingStreet:
			session.Street = text
			session.State = StateAwaitingNumber
			h.sendTelegramMessage(chatID, messages.AwaitingNumber)
		case StateAwaitingNumber:
			session.Number = text
			session.State = StateAwaitingPostcode
			h.sendTelegramMessage(chatID, messages.AwaitingPostcode)
		case StateAwaitingPostcode:
			session.Postcode = text

			items, err := h.garbageService.GetLocationID(r.Context(), session.Street, session.Number, session.Postcode)
			if err != nil {
				session.State = StateNone
				h.sendTelegramMessage(chatID, messages.ErrorFindLocation)
				return
			}

			if len(items) == 0 {
				session.State = StateNone
				h.sendTelegramMessage(chatID, "Nie znaleziono takiej lokalizacji. Wpisz /start, aby spróbować ponownie.")
				return
			}

			// 1. Dynamically build the rows of inline buttons from your items array
			var buttons [][]model.TelegramInlineButton
			for _, item := range items {
				btn := model.TelegramInlineButton{
					Text:         item.FullName,
					CallbackData: fmt.Sprintf("loc_%s", item.AddressPointID), // Prefix prevents collision
				}
				buttons = append(buttons, []model.TelegramInlineButton{btn})
			}

			// Add a fallback cancel button at the bottom
			cancelBtn := model.TelegramInlineButton{Text: messages.CancelButtonLabel, CallbackData: "loc_cancel"}
			buttons = append(buttons, []model.TelegramInlineButton{cancelBtn})

			inlineMenu := &model.TelegramInlineMenu{
				InlineKeyboard: buttons,
			}
			h.sendTelegramMessage(chatID,
				messages.MultipleLocationsPrompt,
				inlineMenu,
			)

			session.State = StateAwaitingLocationConfirmation
		case StateAwaitingLocationConfirmation:
			h.sendTelegramMessage(chatID, messages.AwaitingConfirmationReminder)
		case StateAwaitingSchedule:
			textUpper := strings.ToUpper(strings.TrimSpace(text))

			reDayBefore := regexp.MustCompile(`^W\s*(\d{1,2})(?::00)?$`)
			reMorning := regexp.MustCompile(`^(\d{1,2})(?::00)?$`)

			if matches := reDayBefore.FindStringSubmatch(textUpper); len(matches) > 1 {
				hourInt, _ := strconv.Atoi(matches[1])
				if hourInt < 0 || hourInt > 23 {
					h.sendTelegramMessage(chatID, "Błędna godzina. Podaj poprawną godzinę (od 0 do 23).")
					return
				}
				pref := fmt.Sprintf("day_before_%s", matches[1])

				alreadyHas := false
				for _, p := range session.SelectedPreferences {
					if p == pref {
						alreadyHas = true
						break
					}
				}

				if alreadyHas {
					newMenu := h.buildScheduleKeyboard(session.SelectedPreferences)
					h.sendTelegramMessage(chatID, fmt.Sprintf("Godzina %s jest już wybrana.", formatPreference(pref)), newMenu)
				} else {
					session.SelectedPreferences = append(session.SelectedPreferences, pref)
					newMenu := h.buildScheduleKeyboard(session.SelectedPreferences)
					h.sendTelegramMessage(chatID, fmt.Sprintf("Dodano godzinę: %s", formatPreference(pref)), newMenu)
				}
			} else if matches := reMorning.FindStringSubmatch(textUpper); len(matches) > 1 {
				hourInt, _ := strconv.Atoi(matches[1])
				if hourInt < 0 || hourInt > 23 {
					h.sendTelegramMessage(chatID, "Błędna godzina. Podaj poprawną godzinę (od 0 do 23).")
					return
				}
				pref := fmt.Sprintf("morning_%s", matches[1])

				alreadyHas := false
				for _, p := range session.SelectedPreferences {
					if p == pref {
						alreadyHas = true
						break
					}
				}

				if alreadyHas {
					newMenu := h.buildScheduleKeyboard(session.SelectedPreferences)
					h.sendTelegramMessage(chatID, fmt.Sprintf("Godzina %s jest już wybrana.", formatPreference(pref)), newMenu)
				} else {
					session.SelectedPreferences = append(session.SelectedPreferences, pref)
					newMenu := h.buildScheduleKeyboard(session.SelectedPreferences)
					h.sendTelegramMessage(chatID, fmt.Sprintf("Dodano godzinę: %s", formatPreference(pref)), newMenu)
				}
			} else {
				h.sendTelegramMessage(chatID, messages.AwaitingScheduleReminder)
			}
		default:
			h.sendTelegramMessage(chatID, messages.UnknownCommand)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *TelegramHandler) sendTelegramMessage(chatID int64, text string, markup ...*model.TelegramInlineMenu) {
	var m *model.TelegramInlineMenu
	if len(markup) > 0 {
		m = markup[0]
	}
	if err := h.telegramSvc.SendMessage(context.Background(), chatID, text, m); err != nil {
		log.Printf("TelegramHandler error: %v", err)
	}
}

// sendCallbackAcknowledgment tells Telegram to clear the button loading state
func (h *TelegramHandler) sendCallbackAcknowledgment(callbackQueryID string) {
	if err := h.telegramSvc.AnswerCallbackQuery(context.Background(), callbackQueryID); err != nil {
		log.Printf("TelegramHandler error: %v", err)
	}
}

// sendTelegramEditMessage swaps out the button markup with a plain string confirmation
func (h *TelegramHandler) sendTelegramEditMessage(chatID int64, messageID int64, text string) {
	if err := h.telegramSvc.EditMessageText(context.Background(), chatID, messageID, text); err != nil {
		log.Printf("TelegramHandler error: %v", err)
	}
}

// sendTelegramEditReplyMarkup updates only the inline buttons layout dynamically
func (h *TelegramHandler) sendTelegramEditReplyMarkup(chatID int64, messageID int64, markup *model.TelegramInlineMenu) {
	if err := h.telegramSvc.EditMessageReplyMarkup(context.Background(), chatID, messageID, markup); err != nil {
		log.Printf("TelegramHandler error: %v", err)
	}
}

// buildScheduleKeyboard returns the dynamic keyboard of notification preferences
func (h *TelegramHandler) buildScheduleKeyboard(selected []string) *model.TelegramInlineMenu {
	options := []struct {
		Key  string
		Text string
	}{
		{"day_before_19", messages.OptDayBefore19},
		{"day_before_20", messages.OptDayBefore20},
		{"day_before_21", messages.OptDayBefore21},
		{"morning_7", messages.OptMorning7},
		{"morning_8", messages.OptMorning8},
		{"morning_9", messages.OptMorning9},
		{"morning_10", messages.OptMorning10},
	}

	// Add custom options to the menu so the user can uncheck them
	for _, p := range selected {
		found := false
		for _, o := range options {
			if o.Key == p {
				found = true
				break
			}
		}
		if !found {
			options = append(options, struct{ Key, Text string }{p, formatPreference(p)})
		}
	}

	var keyboard [][]model.TelegramInlineButton
	for _, opt := range options {
		has := false
		for _, s := range selected {
			if s == opt.Key {
				has = true
				break
			}
		}

		icon := "⬜"
		if has {
			icon = "✅"
		}

		keyboard = append(keyboard, []model.TelegramInlineButton{
			{
				Text:         fmt.Sprintf("%s %s", icon, opt.Text),
				CallbackData: "pref_" + opt.Key,
			},
		})
	}

	keyboard = append(keyboard, []model.TelegramInlineButton{
		{
			Text:         messages.DoneButtonLabel,
			CallbackData: "pref_done",
		},
	})

	return &model.TelegramInlineMenu{
		InlineKeyboard: keyboard,
	}
}

func formatPreference(p string) string {
	optionsMap := map[string]string{
		"day_before_19": messages.OptDayBefore19,
		"day_before_20": messages.OptDayBefore20,
		"day_before_21": messages.OptDayBefore21,
		"morning_7":     messages.OptMorning7,
		"morning_8":     messages.OptMorning8,
		"morning_9":     messages.OptMorning9,
		"morning_10":    messages.OptMorning10,
	}
	if val, ok := optionsMap[p]; ok {
		return val
	}
	if strings.HasPrefix(p, "day_before_") {
		hour := strings.TrimPrefix(p, "day_before_")
		return fmt.Sprintf("Dzień wcześniej o %s:00", hour)
	}
	if strings.HasPrefix(p, "morning_") {
		hour := strings.TrimPrefix(p, "morning_")
		return fmt.Sprintf("W dniu wywozu o %s:00", hour)
	}
	return p
}
