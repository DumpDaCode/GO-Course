package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/go-course/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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

// GetUserById returns a user by ID
func (m *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var user models.User

	query := `
		select id, first_name, last_name, email, password, access_level, created_at, update_at 
		from users
		where id = $1
	`

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.AccessLevel,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

// UpdateUser updates the user in database
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	update users 
	set first_name = $1,
	last_name = $2,
	email = $3, 
	access_level = $4,
	updated_at = $5
	`

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	query := `
		select id, password 
		from users
		where email = $1
	`

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&id,
		&hashedPassword,
	)

	if err != nil {
		return 0, "", errors.New("unable to connect to db")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

// AllReservations returns a slice of all reservations
func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var reservations []models.Reservation

	query := `
	select 
		r.id, r.first_name, r.last_name, r.email, 
		r.start_date, r.end_date, r.room_id, 
		r.created_at, r.updated_at, rm.id, 
		rm.room_name
	from reservations r
	left join rooms rm
	on (r.room_id = rm.id)
	order by r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err := rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// AllNewReservations returns a slice of all new reservations
func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var reservations []models.Reservation

	query := `
	select 
		r.id, r.first_name, r.last_name, r.email, 
		r.start_date, r.end_date, r.room_id, 
		r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
	from reservations r
	left join rooms rm
	on (r.room_id = rm.id)
	where processed = 0
	order by r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Procesed,
			&i.Room.ID,
			&i.Room.RoomName,
		)
		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err := rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// GetReservatioById returns a reservation based on id
func (m *postgresDBRepo) GetReservationById(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var reservation models.Reservation

	query := `
	select 
		r.id, r.first_name, r.last_name, r.email, r.phone,
		r.start_date, r.end_date, r.room_id, 
		r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
	from reservations r
	left join rooms rm
	on (r.room_id = rm.id)
	where r.id = $1
	`

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&reservation.ID,
		&reservation.FirstName,
		&reservation.LastName,
		&reservation.Email,
		&reservation.Phone,
		&reservation.StartDate,
		&reservation.EndDate,
		&reservation.RoomID,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&reservation.Procesed,
		&reservation.Room.ID,
		&reservation.Room.RoomName,
	)

	if err != nil {
		return reservation, err
	}

	return reservation, nil
}

// UpdateReservation updates a reservation in database
func (m *postgresDBRepo) UpdateReservation(u models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	update reservations	 
	set first_name = $1,
	last_name = $2,
	email = $3, 
	phone = $4,
	updated_at = $5
	where id = $6
	`

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Phone,
		time.Now(),
		u.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// DeletReservation deletes reservation from databsase
func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	delete 
	from reservations 
	where id =$1
	`

	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	return nil
}

// UpdateProcessed updates processed flag in database
func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	update reservations 
	set processed = $1
	where id =$2
	`

	_, err := m.DB.ExecContext(ctx, query, id, processed)

	if err != nil {
		return err
	}

	return nil
}
