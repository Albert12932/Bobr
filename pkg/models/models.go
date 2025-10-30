package models

import "time"

type Auth struct {
	BookId int `json:"book_id"`
}

type ErrorResponse struct {
	Error   string
	Message string
}

type Student struct {
	Id         int       `json:"id"`
	BookId     int       `json:"book_id"`
	Surname    string    `json:"surname"`
	Name       string    `json:"name"`
	MiddleName string    `json:"middle_name"`
	BirthDate  time.Time `json:"birth_date"`
	Group      string    `json:"group"`
}

type AuthStatus struct {
	Status          string `json:"status"`
	DisplayName     string `json:"display_name"`
	Group           string `json:"group"`
	LinkToken       string `json:"link_token"`
	LinkTokenTtlSec int    `json:"link_token_ttl_sec"`
}

type AuthReq struct {
	Password string `json:"password" binding:"required,min=8"`
	Mail     string `json:"mail"`
}

type LoginRequest struct {
	BookId   int    `json:"book_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResp struct {
	OK   bool `json:"ok"`
	User struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Surname   string `json:"surname"`
	} `json:"user"`
	Session struct {
		Auth struct {
			Token     string    `json:"token"`
			ExpiresAt time.Time `json:"expires_at"`
		} `json:"auth"`
	} `json:"session"`
}
