package dbrepo

import (
	"context"
	"time"

	"github.com/go-course/bookings/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts reservation into the database
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var newId int
	stmt := `
		insert into 
			reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id
	`
	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newId)

	if err != nil {
		return 0, err
	}

	return newId, nil
}

func (m *postgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `
		insert into 
			room_restrictions (start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, $7) returning id
	`
	_, err := m.DB.ExecContext(ctx, stmt,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		res.ReservationID,
		res.RestrictionID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDates returns true if availability exists for roomID and false if no availability exists
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	count := 0

	stmt := `
	SELECT
		count(id)
	FROM
		room_restrictions rr
	WHERE
		room_id = $1 
		and $2 < end_date
		AND $3 > start_date
	`
	err := m.DB.QueryRowContext(ctx, stmt, roomID, start, end).Scan(&count)

	if err != nil {
		return false, err
	}

	if count == 0 {
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var rooms []models.Room
	stmt := `
	SELECT
		r.id,
		r.room_name 
	FROM
		rooms r
	WHERE
		r.id NOT IN (
		SELECT
			rr.room_id
		FROM
			room_restrictions rr
		WHERE
			$1 < rr.end_date
			AND $2 > rr.start_date)
	`
	rows, err := m.DB.QueryContext(ctx, stmt, start, end)
	if err != nil {
		return rooms, err
	}
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return rooms, err
	}
	return rooms, nil
}

func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var rooms models.Room

	query := `
		select id, room_name, created_at, updated_at from rooms where id = $1
	`

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&rooms.ID,
		&rooms.RoomName,
		&rooms.CreatedAt,
		&rooms.UpdatedAt,
	)

	if err != nil {
		return rooms, err
	}

	return rooms, nil
}
