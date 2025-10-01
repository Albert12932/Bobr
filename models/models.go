package models

import "time"

type Auth struct {
	Book_id int `json:"book_id"`
}

type ErrorResponse struct {
	Error   string
	Message string
}

type Student struct {
	Id          int       `json:"id"`
	Book_id     int       `json:"book_id"`
	Surname     string    `json:"surname"`
	Name        string    `json:"name"`
	Middle_name string    `json:"middle_name"`
	Birth_date  time.Time `json:"birth_date"`
	Group       string    `json:"group"`
}

type AuthStatus struct {
	Status             string `json:"status"`
	Display_name       string `json:"display_name"`
	Group              string `json:"group"`
	Link_token         string `json:"link_token"`
	Link_token_ttl_sec int    `json:"link_token_ttl_sec"`
}
