package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"smieci-sms/internal/model"
	"strconv"
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
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// API JSON structs
type Fraction struct {
	IDFrakcja string `json:"id_frakcja"`
	Nazwa     string `json:"nazwa"`
}

type HarmonogramItem struct {
	Data    string   `json:"data"`
	Frakcja Fraction `json:"frakcja"`
}

type SmieciResponseItem struct {
	Adres         string            `json:"adres"`
	Dzielnica     string            `json:"dzielnicy"`
	Harmonogramy  []HarmonogramItem `json:"harmonogramy"`
	HarmonogramyN []HarmonogramItem `json:"harmonogramyN"`
	HarmonogramyZ []HarmonogramItem `json:"harmonogramyZ"`
}

func parseTime(dateStr string) *time.Time {
	if dateStr == "" || dateStr == "1900-01-01" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		log.Printf("Error parsing date '%s': %v", dateStr, err)
		return nil
	}
	return &t
}

func (s *garbageService) FetchSchedulesForLocations(ctx context.Context, locationIDs []int) ([]model.GarbageSchedule, error) {
	var updatedSchedules []model.GarbageSchedule

	baseURL := s.sourceURL

	for _, locID := range locationIDs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		log.Printf("Refreshing schedule from website for location: %d", locID)

		// Construct URL
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid source URL: %w", err)
		}

		q := u.Query()
		q.Set("p_p_id", "portalCKMjunkschedules_WAR_portalCKMjunkschedulesportlet_INSTANCE_o5AIb2mimbRJ")
		q.Set("p_p_lifecycle", "2")
		q.Set("p_p_state", "normal")
		q.Set("p_p_mode", "view")
		q.Set("p_p_resource_id", "ajaxResource")
		q.Set("p_p_cacheability", "cacheLevelPage")
		q.Set("_portalCKMjunkschedules_WAR_portalCKMjunkschedulesportlet_INSTANCE_o5AIb2mimbRJ_addressPointId", strconv.Itoa(locID))
		u.RawQuery = q.Encode()

		// Build HTTP request
		req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add standard headers to look like a browser
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
		}

		var items []SmieciResponseItem
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			return nil, fmt.Errorf("failed to decode JSON response: %w", err)
		}

		// Prepare local timezone's today string (YYYY-MM-DD) for filtering past schedules
		todayStr := time.Now().Local().Format("2006-01-02")

		// We will collect the earliest date >= today for each fraction
		minDates := make(map[string]string)

		if len(items) > 0 {
			// Combine items from harmonogramy and harmonogramyZ
			var allSchedules []HarmonogramItem
			allSchedules = append(allSchedules, items[0].Harmonogramy...)
			allSchedules = append(allSchedules, items[0].HarmonogramyZ...)

			for _, h := range allSchedules {
				code := h.Frakcja.IDFrakcja
				dateStr := h.Data

				// Skip empty dates, placeholders, and past dates
				if dateStr == "" || dateStr == "1900-01-01" || dateStr < todayStr {
					continue
				}

				if currentMin, exists := minDates[code]; !exists || dateStr < currentMin {
					minDates[code] = dateStr
				}
			}
		}

		freshData := model.GarbageSchedule{
			LocationID: locID,
			LastUpdate: time.Now(),
		}

		if d, ok := minDates["ZM"]; ok {
			freshData.DateZmieszane = parseTime(d)
		}
		if d, ok := minDates["OP"]; ok {
			freshData.DatePapier = parseTime(d)
		}
		if d, ok := minDates["MT"]; ok {
			freshData.DatePlastik = parseTime(d)
		}
		if d, ok := minDates["OS"]; ok {
			freshData.DateSzklo = parseTime(d)
		}
		if d, ok := minDates["BK"]; ok {
			freshData.DateBio = parseTime(d)
		}
		if d, ok := minDates["OZ"]; ok {
			freshData.DateZielone = parseTime(d)
		}
		if d, ok := minDates["WG"]; ok {
			freshData.DateGabaryty = parseTime(d)
		}

		updatedSchedules = append(updatedSchedules, freshData)

		// DevOps courtesy rule: wait 500ms between requests so the city doesn't block your IP
		time.Sleep(500 * time.Millisecond)
	}

	return updatedSchedules, nil
}
