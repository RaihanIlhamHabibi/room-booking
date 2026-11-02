package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"example.com/room-booking/internal/model"
	"example.com/room-booking/internal/repository"
)

var ErrInvalid = errors.New("invalid input")

type Service struct{ store repository.Store }

func New(store repository.Store) *Service { return &Service{store: store} }

func (s *Service) Users(ctx context.Context) ([]model.User, error) { return s.store.Users(ctx) }

func (s *Service) Login(ctx context.Context, userID string) (model.User, string, error) {
	if strings.TrimSpace(userID) == "" {
		return model.User{}, "", ErrInvalid
	}
	user, err := s.store.User(ctx, userID)
	if err != nil {
		return model.User{}, "", err
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return model.User{}, "", err
	}
	token := hex.EncodeToString(secret)
	if err := s.store.CreateSession(ctx, token, user.ID, time.Now().Add(7*24*time.Hour)); err != nil {
		return model.User{}, "", err
	}
	return user, token, nil
}

func (s *Service) SessionUser(ctx context.Context, token string) (model.User, error) {
	if token == "" {
		return model.User{}, repository.ErrNotFound
	}
	return s.store.SessionUser(ctx, token)
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.DeleteSession(ctx, token)
}

func (s *Service) Rooms(ctx context.Context) ([]model.Room, error) { return s.store.Rooms(ctx) }
func (s *Service) Room(ctx context.Context, id string) (model.Room, error) {
	return s.store.Room(ctx, id)
}

func (s *Service) UpdateRoomDescription(ctx context.Context, id, description string) (model.Room, error) {
	description = strings.TrimSpace(description)
	if description == "" || len(description) > 500 {
		return model.Room{}, ErrInvalid
	}
	return s.store.UpdateRoomDescription(ctx, id, description)
}

func (s *Service) Bookings(ctx context.Context, roomID string) ([]model.Booking, error) {
	if _, err := s.store.Room(ctx, roomID); err != nil {
		return nil, err
	}
	return s.store.Bookings(ctx, roomID)
}

func (s *Service) CalendarEvents(ctx context.Context, date string) ([]model.CalendarEvent, error) {
	if date == "" {
		return s.store.CalendarEvents(ctx, nil)
	}
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	day, err := time.ParseInLocation("2006-01-02", date, location)
	if err != nil || day.Format("2006-01-02") != date {
		return nil, ErrInvalid
	}
	return s.store.CalendarEvents(ctx, &day)
}
