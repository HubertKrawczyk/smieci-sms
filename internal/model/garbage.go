package model

import "time"

// type GarbageEvent struct {
// 	ID          int64  `json:"id"`
// 	UserID      int64  `json:"user_id"`
// 	City        string `json:"city"`
// 	District    string `json:"district"`
// 	CollectedOn string `json:"collected_on"`
// 	Type        string `json:"type"`
// }

// type GarbageSchedule struct {
// 	// Location   Location   `json:"location"`
// 	Collection []string `json:"collection_dates"`
// }

type GarbageSchedule struct {
	LocationID           int        `json:"location_id" db:"location_id"`
	DateZmieszane        *time.Time `json:"date_zmieszane,omitempty" db:"date_zmieszane"`
	DatePapier           *time.Time `json:"date_papier,omitempty" db:"date_papier"`
	DatePlastik          *time.Time `json:"date_plastik,omitempty" db:"date_plastik"`
	DateSzklo            *time.Time `json:"date_szklo,omitempty" db:"date_szklo"`
	DateBio              *time.Time `json:"date_bio,omitempty" db:"date_bio"`
	DateZielone          *time.Time `json:"date_zielone,omitempty" db:"date_zielone"`
	DateBioRestauracyjne *time.Time `json:"date_bio_restauracyjne,omitempty" db:"date_bio_restauracyjne"`
	DateGabaryty         *time.Time `json:"date_gabaryty,omitempty" db:"date_gabaryty"`
	LastUpdate           time.Time  `json:"last_update" db:"last_update"`
}

func (g GarbageSchedule) HasToday(t time.Time) bool {
	if g.DateZmieszane != nil && sameDate(*g.DateZmieszane, t) {
		return true
	}
	if g.DatePapier != nil && sameDate(*g.DatePapier, t) {
		return true
	}
	if g.DatePlastik != nil && sameDate(*g.DatePlastik, t) {
		return true
	}
	if g.DateSzklo != nil && sameDate(*g.DateSzklo, t) {
		return true
	}
	if g.DateBio != nil && sameDate(*g.DateBio, t) {
		return true
	}
	if g.DateZielone != nil && sameDate(*g.DateZielone, t) {
		return true
	}
	if g.DateBioRestauracyjne != nil && sameDate(*g.DateBioRestauracyjne, t) {
		return true
	}
	if g.DateGabaryty != nil && sameDate(*g.DateGabaryty, t) {
		return true
	}
	return false
}

func sameDate(a, b time.Time) bool {
	a = a.In(b.Location())
	b = b.In(b.Location())
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

type UserGarbageSchedule struct {
	User     UserLocation    `json:"user"`
	Schedule GarbageSchedule `json:"schedule"`
}
