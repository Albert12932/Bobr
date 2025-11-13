package models

import "time"

type AuthBookRequest struct {
	BookId int64 `json:"book_id"`
}

type DeleteUserRequest struct {
	Mail string `json:"mail"`
}

type ErrorResponse struct {
	Error   string
	Message string
}

type DeleteUserResponse struct {
	Deleted bool   `json:"deleted"`
	Mail    string `json:"mail"`
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
	Id         int64     `json:"id" db:"id"`
	BookId     int64     `json:"book_id" db:"book_id"`
	Name       string    `json:"name" db:"name"`
	Surname    string    `json:"surname" db:"surname"`
	MiddleName string    `json:"middle_name" db:"middle_name"`
	BirthDate  time.Time `json:"birth_date" db:"birth_date"`
	Group      string    `json:"student_group" db:"student_group"`
	Password   []byte    `json:"password" db:"password"`
	Mail       string    `json:"mail" db:"mail"`
	RoleLevel  int64     `json:"role_level" db:"role_level"`
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
	UserSubstructure `json:"user"`
	Auth             `json:"auth"`
}

type Auth struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpUnix      int64  `json:"expires_at"`
}

type UserSubstructure struct {
	ID        int64  `json:"id"`
	BookId    int64  `json:"book_id"`
	Mail      string `json:"mail"`
	FirstName string `json:"first_name"`
	RoleLevel int64  `json:"role_level"`
}
type LoginResponse struct {
	UserSubstructure `json:"user"`
	Auth             `json:"auth"`
}

type RefreshTokenResponse struct {
	UserID int64 `json:"user_id"`
	Auth   `json:"auth"`
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

type GetTokensRequest struct {
	UserId    int64
	RoleLevel int64
}

type ProfileResponse struct {
	BookId       int64
	Name         string
	Surname      string
	MiddleName   string
	BirthDate    time.Time
	StudentGroup string
	Mail         string
	RoleLevel    int64
}

type PatchUserRequest struct {
	UserId  int64 `json:"user_id"`
	NewData struct {
		BookId       int64  `json:"book_id,omitempty"`
		Name         string `json:"name,omitempty"`
		Surname      string `json:"surname,omitempty"`
		MiddleName   string `json:"middle_name,omitempty"`
		StudentGroup string `json:"student_group,omitempty"`
		Mail         string `json:"mail,omitempty"`
		RoleLevel    int64  `json:"role_level,omitempty"`
	} `json:"new_data"`
}

type PatchUserResponse struct {
	Successful bool  `json:"successful"`
	UserID     int64 `json:"user_id"`
	New        struct {
		BookId       int64  `json:"book_id,omitempty"`
		Name         string `json:"name,omitempty"`
		Surname      string `json:"surname,omitempty"`
		MiddleName   string `json:"middle_name,omitempty"`
		StudentGroup string `json:"student_group,omitempty"`
		Mail         string `json:"mail,omitempty"`
		RoleLevel    int64  `json:"role_level,omitempty"`
	}
}
