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

type DeleteUserResponse struct {
	Successful bool   `json:"successful"`
	Email      string `json:"email"`
}
type SuccessResponse struct {
	Successful bool   `json:"successful"`
	Message    string `json:"message"`
}

type DeleteEventResponse struct {
	Successful bool `json:"successful"`
	EventID    int  `json:"event_id"`
}

type DeleteCompletedEventResponse struct {
	Successful bool `json:"successful"`
	UserID     int  `json:"user_id"`
	EventID    int  `json:"event_id"`
}

type RefreshTokenResponse struct {
	UserID int64 `json:"user_id"`
	Auth   `json:"auth"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type SetNewPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type Payload struct {
	Sub       int64 `json:"sub"`
	RoleLevel int64 `json:"role_level"`
	Exp       int64 `json:"exp"`
	Iat       int64 `json:"iat"`
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

type UpdateUserRequest struct {
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

type UpdateUserResponse struct {
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

type CreateEventRequest struct {
	Title         string     `json:"title" binding:"required"`
	Description   string     `json:"description"`
	EventTypeCode int        `json:"event_type_code"`
	Points        int        `json:"points"`
	IconUrl       string     `json:"icon_url"`
	EventDate     *time.Time `json:"event_date"`
}

type CreateEventResponse struct {
	EventID       int64     `json:"event_id" db:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	EventTypeCode int       `json:"event_type_code"`
	Points        int       `json:"points"`
	IconUrl       string    `json:"icon_url"`
	EventDate     time.Time `json:"event_date"`
	CreatedAt     time.Time `json:"created"`
}

type UpdateEventRequest struct {
	EventId int64 `json:"event_id"`
	NewData struct {
		Title         string     `json:"title,omitempty"`
		Description   string     `json:"description,omitempty"`
		EventTypeCode int        `json:"event_type_code,omitempty"`
		Points        int        `json:"points,omitempty"`
		IconUrl       string     `json:"icon_url,omitempty"`
		EventDate     *time.Time `json:"event_date,omitempty"`
	} `json:"new_data"`
}

type Event struct {
	EventID       int64     `json:"event_id" db:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	EventTypeCode int       `json:"event_type_code"`
	Points        int       `json:"points"`
	IconUrl       string    `json:"icon_url"`
	EventDate     time.Time `json:"event_date"`
	CreatedAt     time.Time `json:"created"`
}

type CompleteUserEventRequest struct {
	UserId  int64 `json:"user_id"`
	EventId int64 `json:"event_id"`
}

type UserCompletedEvent struct {
	EventID   int64     `json:"event_id" db:"event_id"`
	Completed time.Time `json:"completed_at" db:"completed_at"`
}

type CompletedEvent struct {
	UserId      int64     `json:"user_id"`
	EventId     int64     `json:"event_id"`
	CompletedAt time.Time `json:"completed_at"`
}
