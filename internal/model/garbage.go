package model

type GarbageEvent struct {
    ID         int64  `json:"id"`
    UserID     int64  `json:"user_id"`
    City       string `json:"city"`
    District   string `json:"district"`
    CollectedOn string `json:"collected_on"`
    Type       string `json:"type"`
}

type GarbageSchedule struct {
    Location   Location   `json:"location"`
    Collection []string   `json:"collection_dates"`
}
