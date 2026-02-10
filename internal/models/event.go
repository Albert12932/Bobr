package models

import "time"

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
type UserCompletedEvent struct {
	EventID   int64     `json:"event_id" db:"event_id"`
	Completed time.Time `json:"completed_at" db:"completed_at"`
}
type CompletedEvent struct {
	UserId      int64     `json:"user_id"`
	EventId     int64     `json:"event_id"`
	CompletedAt time.Time `json:"completed_at"`
}
type DeleteEventResponse struct {
	Successful bool  `json:"successful"`
	EventID    int64 `json:"event_id"`
}
type DeleteCompletedEventResponse struct {
	Successful bool  `json:"successful"`
	UserID     int64 `json:"user_id"`
	EventID    int64 `json:"event_id"`
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

type CompleteUserEventRequest struct {
	UserId  int64 `json:"user_id"`
	EventId int64 `json:"event_id"`
}

type CompletedEventsStats struct {
	Hackathons int `json:"hackathons"`
	Articles   int `json:"articles"`
	Olympiads  int `json:"olympiads"`
	Projects   int `json:"projects"`
}

type CompletedEventsFullResponse struct {
	Events []UserCompletedEvent `json:"events"`
	Stats  CompletedEventsStats `json:"stats"`
}

type CreateSuggestRequest struct {
	EventId        int64 `json:"event_id" binding:"required"`
	ExpiresAtHours int64 `json:"expires_at"`
}
