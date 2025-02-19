package models

type Order struct {
	OrderNumber      string `json:"order_number"`
	OrderDate        string `json:"order_date"`
	EstimatedArrival string `json:"estimated_arrival"`
	CarModel         string `json:"car_model"`
	CountryOfOrigin  string `json:"country_of_origin"`
	CurrentLocation  string `json:"current_location"`
	Destination      string `json:"destination"`
}

type User struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	Order1      Order  `json:"order_1"`
	Order2      Order  `json:"order_2"`
}
