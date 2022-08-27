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
