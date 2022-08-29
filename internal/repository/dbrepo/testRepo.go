package dbrepo

import (
	"errors"
	"time"

	"github.com/go-course/bookings/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts reservation into the database
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.RoomID == 2 {
		return 0, errors.New("some error")
	}
	return 1, nil
}

func (m *testDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	if res.RoomID == 100 {
		return errors.New("some error")
	}
	return nil
}

// SearchAvailabilityByDates returns true if availability exists for roomID and false if no availability exists
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	if roomID > 2 {
		return false, errors.New("room not Available")
	}
	return true, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms = []models.Room{
		{
			ID:        0,
			RoomName:  "",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}

	noRoomsDate, _ := time.Parse("2006-01-02", "2022-08-27")

	if start == noRoomsDate {
		return []models.Room{}, nil
	}
	return rooms, nil
}

func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var rooms models.Room
	if id > 2 {
		return rooms, errors.New("some error")
	}
	return rooms, nil
}

// GetUserById returns a user by ID
func (m *testDBRepo) GetUserById(id int) (models.User, error) {
	var user models.User

	return user, nil
}

// UpdateUser updates the user in database
func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

// Authenticate authenticates a user
func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	var id int
	var hashedPassword string

	if email != "rajiv@mkcl.org" {
		return id, hashedPassword, errors.New("some error")
	}
	return id, hashedPassword, nil
}

// AllReservations returns a slice of all reservations
func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

func (m *testDBRepo) GetReservationById(id int) (models.Reservation, error) {

	var reservation models.Reservation
	return reservation, nil
}

func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	return nil
}

func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	return nil
}

func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	var rooms = []models.Room{
		{
			ID:        1,
			RoomName:  "Generals Quarters",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			RoomName:  "Majors Suite",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	return rooms, nil
}

func (m *testDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	var restrictions = []models.RoomRestriction{
		{
			ID:            roomID,
			StartDate:     start,
			EndDate:       end,
			RoomID:        1,
			ReservationID: 1,
			RestrictionID: 1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            roomID,
			StartDate:     start.AddDate(0, 1, 1),
			EndDate:       end.AddDate(0, 1, 1),
			RoomID:        1,
			ReservationID: 0,
			RestrictionID: 2,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	return restrictions, nil
}

func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	return nil
}

func (m *testDBRepo) DeleteBlockByID(id int) error {
	return nil
}
