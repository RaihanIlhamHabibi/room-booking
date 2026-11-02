package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/room-booking/internal/model"
	"example.com/room-booking/internal/repository"
)

type fakeStore struct {
	repository.Store
	updated      string
	createdFor   string
	createdToken string
	eventsDay    *time.Time
}

func (f *fakeStore) User(_ context.Context, id string) (model.User, error) {
	if id != "u-ada" {
		return model.User{}, repository.ErrNotFound
	}
	return model.User{ID: id, Name: "Ada"}, nil
}
func (f *fakeStore) CreateSession(_ context.Context, token, userID string, _ time.Time) error {
	f.createdFor, f.createdToken = userID, token
	return nil
}
func (f *fakeStore) UpdateRoomDescription(_ context.Context, _, description string) (model.Room, error) {
	f.updated = description
	return model.Room{Description: description}, nil
}
func (f *fakeStore) CalendarEvents(_ context.Context, day *time.Time) ([]model.CalendarEvent, error) {
	f.eventsDay = day
	return []model.CalendarEvent{}, nil
}

func TestLoginUsesExistingUserAndOpaqueSession(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)
	if _, _, err := svc.Login(context.Background(), "unknown"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("unknown user: %v", err)
	}
	user, token, err := svc.Login(context.Background(), "u-ada")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != "u-ada" || store.createdFor != user.ID || token != store.createdToken || len(token) != 64 {
		t.Fatalf("session was not stored for selected user")
	}
}

func TestDescriptionValidationHappensBeforePersistence(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)
	if _, err := svc.UpdateRoomDescription(context.Background(), "r-1", "   "); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty description: %v", err)
	}
	if store.updated != "" {
		t.Fatal("invalid description reached repository")
	}
	room, err := svc.UpdateRoomDescription(context.Background(), "r-1", "  New notes  ")
	if err != nil || room.Description != "New notes" {
		t.Fatalf("trimmed update: %+v %v", room, err)
	}
}

func TestCalendarDateUsesApplicationTimeZone(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)
	if _, err := svc.CalendarEvents(context.Background(), "2030-02-30"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid date: %v", err)
	}
	if store.eventsDay != nil {
		t.Fatal("invalid date reached repository")
	}
	if _, err := svc.CalendarEvents(context.Background(), "2030-05-13"); err != nil {
		t.Fatal(err)
	}
	if store.eventsDay == nil || store.eventsDay.Hour() != 0 || store.eventsDay.Format("-07:00") != "+07:00" {
		t.Fatalf("wrong day boundary: %v", store.eventsDay)
	}
}
