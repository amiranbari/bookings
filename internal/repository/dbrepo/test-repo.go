package dbrepo

import (
	"errors"
	"github.com/amiranbari/bookings/pkg/models"
	"time"
)

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	//return error if room id eq 2
	if res.RoomId == 2 {
		return 0, errors.New("Some error!")
	}
	return 1, nil
}

func (m *testDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	//return error if room id eq 100
	if res.RoomId == 100 {
		return errors.New("Some error!")
	}
	return nil
}

func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	if start.Format("2006-01-02") == "2040-01-01" {
		return rooms, errors.New("Some error!")
	}

	if start.Format("2006-01-02") == "2040-02-01" {
		rooms = append(rooms, models.Room{})
		return rooms, nil
	}

	return rooms, nil
}

func (m *testDBRepo) GetRoomById(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("Some error!")
	}
	return room, nil
}

func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var user models.User
	return user, nil
}

func (m *testDBRepo) Authenticate(email, password string) (int, string, error) {
	return 1, "", nil

}
