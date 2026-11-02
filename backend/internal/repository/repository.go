package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"example.com/room-booking/internal/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

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
	CalendarEvents(context.Context, *time.Time) ([]model.CalendarEvent, error)
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

func (s *GormStore) CalendarEvents(ctx context.Context, day *time.Time) ([]model.CalendarEvent, error) {
	rows := make([]model.CalendarEvent, 0)
	query := s.DB.WithContext(ctx).Table("calendar_events")
	if day != nil {
		query = query.Where("starts_at < ? AND ends_at > ?", day.AddDate(0, 0, 1), day)
	}
	err := query.Order("starts_at, id").Find(&rows).Error
	return rows, err
}
