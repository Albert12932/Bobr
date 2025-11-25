package models

import "time"

type Student struct {
	Id           int64     `json:"id"`
	BookId       int64     `json:"book_id"`
	Surname      string    `json:"surname"`
	Name         string    `json:"name"`
	MiddleName   string    `json:"middle_name"`
	BirthDate    time.Time `json:"birth_date"`
	StudentGroup string    `json:"student_group" db:"student_group"`
}

type User struct {
	Id           int64     `json:"id" db:"id"`
	BookId       int64     `json:"book_id" db:"book_id"`
	Name         string    `json:"name" db:"name"`
	Surname      string    `json:"surname" db:"surname"`
	MiddleName   string    `json:"middle_name" db:"middle_name"`
	BirthDate    time.Time `json:"birth_date" db:"birth_date"`
	StudentGroup string    `json:"student_group" db:"student_group"`
	Password     []byte    `json:"password" db:"password"`
	Email        string    `json:"email" db:"email"`
	RoleLevel    int64     `json:"role_level" db:"role_level"`
}
type UserSubstructure struct {
	ID           int64  `json:"id"`
	BookId       int64  `json:"book_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	RoleLevel    int64  `json:"role_level"`
	StudentGroup string `json:"student_group"`
}

type ProfileResponse struct {
	BookId       int64     `json:"book_id"`
	Name         string    `json:"name"`
	Surname      string    `json:"surname"`
	MiddleName   string    `json:"middle_name"`
	BirthDate    time.Time `json:"birth_date"`
	StudentGroup string    `json:"student_group"`
	Email        string    `json:"email"`
	RoleLevel    int64     `json:"role_level"`
	TotalPoints  int64     `json:"total_points"`
}

type DeleteUserRequest struct {
	Email string `json:"email"`
}
type DeleteUserResponse struct {
	Successful bool   `json:"successful"`
	Email      string `json:"email"`
}

type UserUpdateData struct {
	BookId       int64  `json:"book_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Surname      string `json:"surname,omitempty"`
	MiddleName   string `json:"middle_name,omitempty"`
	StudentGroup string `json:"student_group,omitempty"`
	Email        string `json:"email,omitempty"`
	RoleLevel    int64  `json:"role_level,omitempty"`
}
type UpdateUserRequest struct {
	UserId  int64          `json:"user_id"`
	NewData UserUpdateData `json:"new_data"`
}
type UpdateUserResponse struct {
	Successful bool           `json:"successful"`
	UserID     int64          `json:"user_id"`
	New        UserUpdateData `json:"new_data"`
}

type UserRating struct {
	UserId   int64 `json:"user_id"`
	Points   int64 `json:"points"`
	Position int64 `json:"position"`
}
type UserWithPoints struct {
	UserId      int64  `json:"user_id"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	TotalPoints int64  `json:"total_points"`
	Position    int64  `json:"position"`
}
