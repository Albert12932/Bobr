package models

import "time"

type AuthBookRequest struct {
	BookId int64 `json:"book_id"`
}

type DeleteUserRequest struct {
	Email string `json:"email"`
}

type ErrorResponse struct {
	Error   string
	Message string
}

type DeleteUserResponse struct {
	Deleted bool   `json:"deleted"`
	Email   string `json:"email"`
}

type Student struct {
	Id         int64     `json:"id"`
	BookId     int64     `json:"book_id"`
	Surname    string    `json:"surname"`
	Name       string    `json:"name"`
	MiddleName string    `json:"middle_name"`
	BirthDate  time.Time `json:"birth_date"`
	Group      string    `json:"group" db:"student_group"`
}

type User struct {
	Id         int64     `json:"id"`
	BookId     int64     `json:"book_id"`
	Surname    string    `json:"surname"`
	Name       string    `json:"name"`
	MiddleName string    `json:"middle_name"`
	BirthDate  time.Time `json:"birth_date"`
	Group      string    `json:"student_group" db:"student_group"`
	Password   []byte    `json:"password"`
	Email      string    `json:"email"`
	RoleLevel  int64     `json:"role_level"`
}

type AuthStatus struct {
	Status          string `json:"status"`
	DisplayName     string `json:"display_name"`
	Group           string `json:"group"`
	LinkToken       string `json:"link_token"`
	LinkTokenTtlSec int64  `json:"link_token_ttl_sec"`
}

type RegisterRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RegisterResponse struct {
	OK               bool `json:"ok"`
	UserSubstructure `json:"user"`
}

type Session struct {
	Auth struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		ExpiresAt    time.Time `json:"expires_at"`
	} `json:"auth"`
}

type UserSubstructure struct {
	ID        int64  `json:"id"`
	BookId    int64  `json:"book_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	RoleLevel int64  `json:"role_level"`
	Group     string `json:"group"`
}
type LoginResponse struct {
	UserSubstructure `json:"user"`
	Session          `json:"session"`
}

type RefreshTokenResponse struct {
	UserID  int64 `json:"user_id"`
	Session `json:"session"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordResponse struct {
	OK      bool   `json:"ok"`
	Email   string `json:"email"`
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

type Payload struct {
	Sub       int64 `json:"sub"`
	RoleLevel int64 `json:"role_level"`
	Exp       int64 `json:"exp"`
	Iat       int64 `json:"iat"`
}

type SetRoleRequest struct {
	UserId    int64 `json:"user_id"`
	RoleLevel int64 `json:"role_level"`
}
type SetRoleResponse struct {
	Successful bool  `json:"successful"`
	UserID     int64 `json:"user_id"`
	RoleLevel  int64 `json:"role_level"`
}
