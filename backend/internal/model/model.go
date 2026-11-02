package model

import "time"

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Room struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Capacity    int      `json:"capacity"`
	Amenities   []string `json:"amenities"`
	Description string   `json:"description"`
}

type Booking struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"roomId"`
	OrganizerID string    `json:"organizerId"`
	StartsAt    time.Time `json:"startsAt"`
	EndsAt      time.Time `json:"endsAt"`
	Purpose     string    `json:"purpose"`
}

type CalendarEvent struct {
	ID       string    `json:"id"`
	UserID   string    `json:"userId"`
	Title    string    `json:"title"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
}
