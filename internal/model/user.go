package model

type Address struct {
    Street   string `json:"street"`
    Number   string `json:"number"`
    Postcode string `json:"postcode"`
}

type Location struct {
    ID int64 `json:"id"`
}

type User struct {
    ID         int64 `json:"id"`
    Name       string `json:"name"`
    Phone      string `json:"phone"`
    LocationID int64  `json:"location_id"`
}

type UserLocationRequest struct {
    Name    string  `json:"name"`
    Phone   string  `json:"phone"`
    Address Address `json:"address"`
}
