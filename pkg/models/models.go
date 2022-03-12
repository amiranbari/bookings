package models

import (
	"time"
)

// User is the user model
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Room is the Rooms model
type Room struct {
	ID        int
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Restriction is the Restrictions model
type Restriction struct {
	ID              int
	Title           string
	RestrictionName time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Reservation is the Reservations model
type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomId    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room      Room
}

// RoomRestriction is the RoomRestrictions model
type RoomRestriction struct {
	ID            int
	RoomId        int
	ReservationId int
	RestrictionId int
	StartDate     time.Time
	EndDate       time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

type MailData struct {
	To      string
	From    string
	Subject string
	Content string
}
