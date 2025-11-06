package models

import "time"

type AuthBookRequest struct {
	BookId int `json:"book_id"`
}

type ErrorResponse struct {
	Error   string
	Message string
}

type DeleteUserResponse struct {
	Deleted bool `json:"deleted"`
	BookId  int  `json:"book_id"`
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

type User struct {
	Id         int       `json:"id"`
	BookId     int       `json:"book_id"`
	Surname    string    `json:"surname"`
	Name       string    `json:"name"`
	MiddleName string    `json:"middle_name"`
	BirthDate  time.Time `json:"birth_date"`
	Group      string    `json:"group"`
	Password   []byte    `json:"password"`
	Mail       string    `json:"mail"`
}

type AuthStatus struct {
	Status          string `json:"status"`
	DisplayName     string `json:"display_name"`
	Group           string `json:"group"`
	LinkToken       string `json:"link_token"`
	LinkTokenTtlSec int    `json:"link_token_ttl_sec"`
}

type RegisterRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password"`
	Mail     string `json:"mail"`
}

type LoginRequest struct {
	Mail     string `json:"mail" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RegisterResponse struct {
	OK   bool `json:"ok"`
	User struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Surname   string `json:"surname"`
	}
}

type Session struct {
	Auth struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		ExpiresAt    time.Time `json:"expires_at"`
	} `json:"auth"`
}

type LoginResponse struct {
	User struct {
		ID        int64  `json:"id"`
		Mail      string `json:"mail"`
		FirstName string `json:"first_name"`
		Surname   string `json:"surname"`
	} `json:"user"`
	Session `json:"session"`
}

type RefreshTokenResponse struct {
	UserID  int64 `json:"user_id"`
	Session `json:"session"`
}

type ResetPasswordRequest struct {
	Mail string `json:"mail" binding:"required"`
}

type ResetPasswordResponse struct {
	OK      bool   `json:"ok"`
	Mail    string `json:"mail"`
	Message string `json:"message" default:"If the email is registered, a password reset link has been sent."`
}

type SetNewPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type SetNewPasswordResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}
