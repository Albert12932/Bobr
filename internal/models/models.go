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
	Id         int64     `json:"id" db:"id"`
	BookId     int64     `json:"book_id" db:"book_id"`
	Name       string    `json:"name" db:"name"`
	Surname    string    `json:"surname" db:"surname"`
	MiddleName string    `json:"middle_name" db:"middle_name"`
	BirthDate  time.Time `json:"birth_date" db:"birth_date"`
	Group      string    `json:"student_group" db:"student_group"`
	Password   []byte    `json:"password" db:"password"`
	Email      string    `json:"email" db:"email"`
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
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	RoleLevel int64  `json:"role_level"`
	Group     string `json:"student_group"`
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
	Email        string
	RoleLevel    int64
}

type PatchUserRequest struct {
	UserId  int64 `json:"user_id"`
	NewData struct {
		BookId       int64  `json:"book_id,omitempty" `
		Name         string `json:"name,omitempty" example:""`
		Surname      string `json:"surname,omitempty" example:""`
		MiddleName   string `json:"middle_name,omitempty" example:""`
		StudentGroup string `json:"student_group,omitempty" example:""`
		Email        string `json:"email,omitempty" example:""`
		RoleLevel    int64  `json:"role_level,omitempty"`
	} `json:"new_data"`
}

type PatchUserResponse struct {
	Successful bool  `json:"successful"`
	UserID     int64 `json:"user_id"`
	New        struct {
		BookId       int64  `json:"book_id,omitempty"`
		Name         string `json:"name,omitempty" example:""`
		Surname      string `json:"surname,omitempty" example:""`
		MiddleName   string `json:"middle_name,omitempty" example:""`
		StudentGroup string `json:"student_group,omitempty" example:""`
		Email        string `json:"email,omitempty" example:""`
		RoleLevel    int64  `json:"role_level,omitempty"`
	}
}
