package repository

import (
	"github.com/amiranbari/bookings/pkg/models"
	"time"
)

type DatabaseRepo interface {
	InsertReservation(res models.Reservation) (int, error)

	InsertRoomRestriction(r models.RoomRestriction) error

	SearchAvailabilityByDates(start, end time.Time, roomID int) (bool, error)
}
