package dbrepo

import (
	"context"
	"errors"
	"github.com/amiranbari/bookings/pkg/models"
	"golang.org/x/crypto/bcrypt"
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

func (m *PostgresDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `select id, title, created_at, updated_at from rooms where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&room.ID, &room.Title, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return room, err
	}

	return room, nil
}

func (m *PostgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User

	query := `select * from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.ID, user.FirstName, user.LastName, user.Email, user.Password, user.AccessLevel, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (m *PostgresDBRepo) Authenticate(email, password string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return 0, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil

}

func (m *PostgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
				select r.id, r.first_name, r.last_name,
				       r.email, r.phone, r.start_date, r.end_date, r.created_at, r.updated_at, r.processed, rm.title 
				from reservation r
				left join rooms rm
				on rm.id = r.room_id
				order by r.id, r.start_date desc, r.processed ASC 
				`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	for rows.Next() {
		var i models.Reservation
		err = rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.Title,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil

}

func (m *PostgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
				select r.id, r.first_name, r.last_name,
				       r.email, r.phone, r.start_date, r.end_date, r.created_at, r.updated_at, r.processed, rm.title 
				from reservation r
				left join rooms rm
				on rm.id = r.room_id
				where processed = 0
				order by r.start_date desc 
				`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	for rows.Next() {
		var i models.Reservation
		err = rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.Title,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil

}
