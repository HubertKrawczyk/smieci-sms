package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSchedulesForLocations(t *testing.T) {
	mockResponse := []SmieciResponseItem{
		{
			Adres:     "AAABBBstreet name 6",
			Dzielnica: "dzielnica agbc",
			Harmonogramy: []HarmonogramItem{
				{
					Data: "2026-06-12",
					Frakcja: Fraction{
						IDFrakcja: "OP",
						Nazwa:     "opakowania z papieru i tektury",
					},
				},
				{
					Data: "2026-06-12",
					Frakcja: Fraction{
						IDFrakcja: "MT",
						Nazwa:     "zmieszane odpady opakowaniowe",
					},
				},
				{
					Data: "2026-07-01",
					Frakcja: Fraction{
						IDFrakcja: "OS",
						Nazwa:     "opakowania ze szkła",
					},
				},
				{
					Data: "1900-01-01",
					Frakcja: Fraction{
						IDFrakcja: "OZ",
						Nazwa:     "odpady ulegające biodegradacji",
					},
				},
				{
					Data: "2026-06-17",
					Frakcja: Fraction{
						IDFrakcja: "BK",
						Nazwa:     "odpady kuchenne ulegające biodegradacji",
					},
				},
				{
					Data: "1900-01-01",
					Frakcja: Fraction{
						IDFrakcja: "WG",
						Nazwa:     "odpady wielkogabarytowe",
					},
				},
				{
					Data: "2026-06-12",
					Frakcja: Fraction{
						IDFrakcja: "ZM",
						Nazwa:     "niesegregowane (zmieszane) odpady komunalne",
					},
				},
			},
			HarmonogramyN: []HarmonogramItem{},
			HarmonogramyZ: []HarmonogramItem{
				{
					Data: "2026-06-12",
					Frakcja: Fraction{
						IDFrakcja: "OP",
						Nazwa:     "opakowania z papieru i tektury",
					},
				},
				{
					Data: "2026-06-12",
					Frakcja: Fraction{
						IDFrakcja: "MT",
						Nazwa:     "zmieszane odpady opakowaniowe",
					},
				},
				{
					Data: "2026-07-01",
					Frakcja: Fraction{
						IDFrakcja: "OS",
						Nazwa:     "opakowania ze szkła",
					},
				},
				{
					Data: "1900-01-01",
					Frakcja: Fraction{
						IDFrakcja: "OZ",
						Nazwa:     "odpady ulegające biodegradacji",
					},
				},
				{
					Data: "2026-06-17",
					Frakcja: Fraction{
						IDFrakcja: "BK",
						Nazwa:     "odpady kuchenne ulegające biodegradacji",
					},
				},
				{
					Data: "1900-01-01",
					Frakcja: Fraction{
						IDFrakcja: "WG",
						Nazwa:     "odpady wielkogabarytowe",
					},
				},
				{
					Data: "2026-06-12",
					Frakcja: Fraction{
						IDFrakcja: "ZM",
						Nazwa:     "niesegregowane (zmieszane) odpady komunalne",
					},
				},
			},
		},
	}

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		q := r.URL.Query()
		if q.Get("p_p_id") != "portalCKMjunkschedules_WAR_portalCKMjunkschedulesportlet_INSTANCE_o5AIb2mimbRJ" {
			t.Errorf("expected p_p_id to be portalCKMjunkschedules_WAR_portalCKMjunkschedulesportlet_INSTANCE_o5AIb2mimbRJ, got %s", q.Get("p_p_id"))
		}
		if q.Get("_portalCKMjunkschedules_WAR_portalCKMjunkschedulesportlet_INSTANCE_o5AIb2mimbRJ_addressPointId") != "1234" {
			t.Errorf("expected addressPointId to be 1234, got %s", q.Get("_portalCKMjunkschedules_WAR_portalCKMjunkschedulesportlet_INSTANCE_o5AIb2mimbRJ_addressPointId"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer ts.Close()

	service := NewGarbageService(ts.URL)
	schedules, err := service.FetchSchedulesForLocations(context.Background(), []int{1234})
	if err != nil {
		t.Fatalf("FetchSchedulesForLocations failed: %v", err)
	}

	if len(schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedules))
	}

	sched := schedules[0]
	if sched.LocationID != 1234 {
		t.Errorf("expected location ID 1234, got %d", sched.LocationID)
	}

	// Helper to check expected dates
	checkExpectedDate := func(fieldName string, date *time.Time, expectedStr string) {
		if expectedStr == "" {
			if date != nil {
				t.Errorf("expected field %s to be nil, got %v", fieldName, date)
			}
			return
		}
		if date == nil {
			t.Errorf("expected field %s to be %s, got nil", fieldName, expectedStr)
			return
		}
		actualStr := date.Format("2006-01-02")
		if actualStr != expectedStr {
			t.Errorf("expected field %s to be %s, got %s", fieldName, expectedStr, actualStr)
		}
	}

	// Since current time is 2026-06-12 in metadata:
	// "2026-06-12" >= today ("2026-06-12") -> true
	// "2026-07-01" >= today -> true
	// "1900-01-01" -> placeholder, ignored
	checkExpectedDate("DatePapier", sched.DatePapier, "2026-06-12")
	checkExpectedDate("DatePlastik", sched.DatePlastik, "2026-06-12")
	checkExpectedDate("DateSzklo", sched.DateSzklo, "2026-07-01")
	checkExpectedDate("DateBio", sched.DateBio, "2026-06-17")
	checkExpectedDate("DateZmieszane", sched.DateZmieszane, "2026-06-12")
	checkExpectedDate("DateZielone", sched.DateZielone, "")
	checkExpectedDate("DateGabaryty", sched.DateGabaryty, "")
}
