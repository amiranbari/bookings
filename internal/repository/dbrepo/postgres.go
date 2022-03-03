package dbrepo

import (
	"context"
	"github.com/amiranbari/bookings/pkg/models"
	"time"
)

func (m *PostgresDBRepo) InsertReservation(res models.Reservation) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newId int

	stmt := `INSERT INTO reservation (first_name, last_name, email, phone, start_date, end_date, room_id, created_at ,updated_at)
	       VALUES
	       ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomId,
		time.Now(),
		time.Now(),
	).Scan(&newId)

	if err != nil {
		return 0, err
	}
	return newId, nil
}

func (m *PostgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO room_restrictions (room_id, reservation_id, restriction_id, start_date, end_date, created_at ,updated_at)
	       VALUES
	       ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		res.RoomId,
		res.ReservationId,
		res.RestrictionId,
		res.StartDate,
		res.EndDate,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}
	return nil
}

func (m *PostgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
			select 
				count(id)
			from
				room_restrictions
			where 
			    room_id = $1
			    and
				$2 < end_date and $3 > start_date`

	var numRows int

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)

	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

func (m *PostgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
				select r.id, r.title 
			from 
				rooms r
				where r.id not in 
				(select rr.room_id from room_restrictions rr where $1 < end_date and $2 > start_date) 
			`

	rows, err := m.DB.QueryContext(ctx, query, start, end)

	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.Title)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}
