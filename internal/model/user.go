package model

// type Address struct {
//     Street   string `json:"street"`
//     Number   string `json:"number"`
//     Postcode string `json:"postcode"`
// }

// type Location struct {
//     ID int64 `json:"id"`
// }

type UserLocation struct {
	ID                   int64    `json:"id"`
	ChatID               int64    `json:"chat_id"`
	Name                 string   `json:"name"`
	Phone                string   `json:"phone"`
	LocationID           string   `json:"location_id"`
	AddressName          string   `json:"address_name"`
	NotificationSettings []string `json:"notification_settings"`
}

type UserLocationRequest struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	AddressName string `json:"address_name"`
	LocationID  string `json:"location_id"`
}

type FindLocationIDRequest struct {
	Street   string `json:"street"`
	Number   string `json:"number"`
	Postcode string `json:"postcode"`
}
