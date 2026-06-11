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
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	LocationID  int64  `json:"location_id"`
	AddressName string `json:"address_name"`
}

type UserLocationRequest struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	AddressName string `json:"address_name"`
	LocationID  int64  `json:"location_id"`
}

type FindLocationIDRequest struct {
	Street   string `json:"street"`
	Number   string `json:"number"`
	Postcode string `json:"postcode"`
}
