package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"example.com/room-booking/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("booking conflicts with an existing reservation")
)

type Store interface {
	Users(context.Context) ([]model.User, error)
	User(context.Context, string) (model.User, error)
	CreateSession(context.Context, string, string, time.Time) error
	SessionUser(context.Context, string) (model.User, error)
	DeleteSession(context.Context, string) error
	Rooms(context.Context) ([]model.Room, error)
	Room(context.Context, string) (model.Room, error)
	UpdateRoomDescription(context.Context, string, string) (model.Room, error)
	Bookings(context.Context, string) ([]model.Booking, error)
	CreateBooking(context.Context, model.Booking) error
	AvailableRooms(context.Context, time.Time, time.Time, int) ([]model.Room, error)
	ScheduleData(context.Context, time.Time, []string) (ScheduleData, error)
	CalendarEvents(context.Context, *time.Time) ([]model.CalendarEvent, error)
}

type ScheduleData struct {
	CalendarEvents []model.CalendarEvent
	Bookings       []model.Booking
}

type GormStore struct{ DB *gorm.DB }

func New(db *gorm.DB) Store { return &GormStore{DB: db} }

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *GormStore) Users(ctx context.Context) ([]model.User, error) {
	rows := make([]model.User, 0)
	err := s.DB.WithContext(ctx).Table("users").Order("name").Find(&rows).Error
	return rows, err
}

func (s *GormStore) User(ctx context.Context, id string) (model.User, error) {
	var user model.User
	err := s.DB.WithContext(ctx).Table("users").Where("id = ?", id).First(&user).Error
	return user, notFound(err)
}

func (s *GormStore) CreateSession(ctx context.Context, token, userID string, expires time.Time) error {
	return s.DB.WithContext(ctx).Exec("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)", token, userID, expires).Error
}

func (s *GormStore) SessionUser(ctx context.Context, token string) (model.User, error) {
	var user model.User
	err := s.DB.WithContext(ctx).Table("users").Joins("JOIN sessions ON sessions.user_id = users.id").Where("sessions.token = ? AND sessions.expires_at > ?", token, time.Now()).Select("users.id, users.name, users.email").First(&user).Error
	return user, notFound(err)
}

func (s *GormStore) DeleteSession(ctx context.Context, token string) error {
	return s.DB.WithContext(ctx).Exec("DELETE FROM sessions WHERE token = ?", token).Error
}

type roomRow struct {
	ID, Name, Description string
	Capacity              int
	Amenities             []byte
}

func decodeRoom(row roomRow) (model.Room, error) {
	room := model.Room{ID: row.ID, Name: row.Name, Capacity: row.Capacity, Description: row.Description}
	if err := json.Unmarshal(row.Amenities, &room.Amenities); err != nil {
		return model.Room{}, err
	}
	return room, nil
}

func (s *GormStore) Rooms(ctx context.Context) ([]model.Room, error) {
	var rows []roomRow
	if err := s.DB.WithContext(ctx).Table("rooms").Order("name").Find(&rows).Error; err != nil {
		return nil, err
	}
	rooms := make([]model.Room, 0, len(rows))
	for _, row := range rows {
		room, err := decodeRoom(row)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (s *GormStore) Room(ctx context.Context, id string) (model.Room, error) {
	var row roomRow
	err := s.DB.WithContext(ctx).Table("rooms").Where("id = ?", id).First(&row).Error
	if err != nil {
		return model.Room{}, notFound(err)
	}
	return decodeRoom(row)
}

func (s *GormStore) UpdateRoomDescription(ctx context.Context, id, description string) (model.Room, error) {
	result := s.DB.WithContext(ctx).Table("rooms").Where("id = ?", id).Update("description", description)
	if result.Error != nil {
		return model.Room{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.Room{}, ErrNotFound
	}
	return s.Room(ctx, id)
}

func (s *GormStore) Bookings(ctx context.Context, roomID string) ([]model.Booking, error) {
	rows := make([]model.Booking, 0)
	err := s.DB.WithContext(ctx).Table("bookings").Where("room_id = ?", roomID).Order("starts_at, id").Find(&rows).Error
	return rows, err
}

func (s *GormStore) CreateBooking(ctx context.Context, booking model.Booking) error {
	err := s.DB.WithContext(ctx).Create(&booking).Error
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23P01" {
		return ErrConflict
	}
	return err
}

func (s *GormStore) AvailableRooms(ctx context.Context, startsAt, endsAt time.Time, attendeeCount int) ([]model.Room, error) {
	var rows []roomRow
	err := s.DB.WithContext(ctx).Table("rooms").
		Where("capacity >= ?", attendeeCount).
		Where("NOT EXISTS (SELECT 1 FROM bookings WHERE bookings.room_id = rooms.id AND bookings.starts_at < ? AND bookings.ends_at > ?)", endsAt, startsAt).
		Order("name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	rooms := make([]model.Room, 0, len(rows))
	for _, row := range rows {
		room, err := decodeRoom(row)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (s *GormStore) ScheduleData(ctx context.Context, day time.Time, attendeeIDs []string) (ScheduleData, error) {
	data := ScheduleData{
		CalendarEvents: make([]model.CalendarEvent, 0),
		Bookings:       make([]model.Booking, 0),
	}
	dayEnd := day.AddDate(0, 0, 1)
	if err := s.DB.WithContext(ctx).Table("calendar_events").
		Where("user_id IN ?", attendeeIDs).
		Where("starts_at < ? AND ends_at > ?", dayEnd, day).
		Order("starts_at, id").
		Find(&data.CalendarEvents).Error; err != nil {
		return ScheduleData{}, err
	}
	if err := s.DB.WithContext(ctx).Table("bookings").
		Where("starts_at < ? AND ends_at > ?", dayEnd, day).
		Order("starts_at, id").
		Find(&data.Bookings).Error; err != nil {
		return ScheduleData{}, err
	}
	return data, nil
}

func (s *GormStore) CalendarEvents(ctx context.Context, day *time.Time) ([]model.CalendarEvent, error) {
	rows := make([]model.CalendarEvent, 0)
	query := s.DB.WithContext(ctx).Table("calendar_events")
	if day != nil {
		query = query.Where("starts_at < ? AND ends_at > ?", day.AddDate(0, 0, 1), day)
	}
	err := query.Order("starts_at, id").Find(&rows).Error
	return rows, err
}
