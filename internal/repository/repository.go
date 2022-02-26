package repository

import "github.com/amiranbari/bookings/pkg/models"

type DatabaseRepo interface {
	InsertReservation(res models.Reservation) error
}
