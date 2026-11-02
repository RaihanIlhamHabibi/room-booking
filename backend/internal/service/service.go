package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

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

func parseInterval(date, startTime, endTime string) (time.Time, time.Time, error) {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	startsAt, err := time.ParseInLocation("2006-01-02 15:04", date+" "+startTime, location)
	if err != nil || startsAt.Format("2006-01-02 15:04") != date+" "+startTime {
		return time.Time{}, time.Time{}, ErrInvalid
	}
	endsAt, err := time.ParseInLocation("2006-01-02 15:04", date+" "+endTime, location)
	if err != nil || endsAt.Format("2006-01-02 15:04") != date+" "+endTime || !endsAt.After(startsAt) {
		return time.Time{}, time.Time{}, ErrInvalid
	}
	return startsAt, endsAt, nil
}

func (s *Service) CreateBooking(ctx context.Context, organizerID, roomID, date, startTime, endTime, purpose string) (model.Booking, error) {
	startsAt, endsAt, err := parseInterval(date, startTime, endTime)
	if err != nil {
		return model.Booking{}, err
	}
	purpose = strings.TrimSpace(purpose)
	if purpose == "" || utf8.RuneCountInString(purpose) > 200 {
		return model.Booking{}, ErrInvalid
	}
	if _, err := s.store.Room(ctx, roomID); err != nil {
		return model.Booking{}, err
	}
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return model.Booking{}, err
	}
	booking := model.Booking{
		ID:          "b-" + hex.EncodeToString(idBytes),
		RoomID:      roomID,
		OrganizerID: organizerID,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Purpose:     purpose,
	}
	if err := s.store.CreateBooking(ctx, booking); err != nil {
		return model.Booking{}, err
	}
	return booking, nil
}

func (s *Service) AvailableRooms(ctx context.Context, date, startTime, endTime string, attendeeCount int) ([]model.Room, error) {
	if attendeeCount < 1 {
		return nil, ErrInvalid
	}
	startsAt, endsAt, err := parseInterval(date, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return s.store.AvailableRooms(ctx, startsAt, endsAt, attendeeCount)
}

func (s *Service) Schedule(ctx context.Context, organizerID, date string, durationMinutes int, attendeeIDs []string) ([]model.ScheduleOption, error) {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}
	day, err := time.ParseInLocation("2006-01-02", date, location)
	if err != nil || day.Format("2006-01-02") != date {
		return nil, ErrInvalid
	}
	if durationMinutes < 15 || durationMinutes > 480 || durationMinutes%15 != 0 {
		return nil, ErrInvalid
	}

	users, err := s.store.Users(ctx)
	if err != nil {
		return nil, err
	}
	knownUsers := make(map[string]struct{}, len(users))
	for _, user := range users {
		knownUsers[user.ID] = struct{}{}
	}
	if _, ok := knownUsers[organizerID]; !ok {
		return nil, repository.ErrNotFound
	}

	uniqueIDs := make(map[string]struct{}, len(attendeeIDs)+1)
	uniqueIDs[organizerID] = struct{}{}
	for _, id := range attendeeIDs {
		if _, ok := knownUsers[id]; !ok {
			return nil, ErrInvalid
		}
		uniqueIDs[id] = struct{}{}
	}
	if len(uniqueIDs) < 2 {
		return nil, ErrInvalid
	}
	attendees := make([]string, 0, len(uniqueIDs))
	for id := range uniqueIDs {
		attendees = append(attendees, id)
	}

	rooms, err := s.store.Rooms(ctx)
	if err != nil {
		return nil, err
	}
	data, err := s.store.ScheduleData(ctx, day, attendees)
	if err != nil {
		return nil, err
	}

	attendeeSet := make(map[string]struct{}, len(attendees))
	for _, id := range attendees {
		attendeeSet[id] = struct{}{}
	}
	attendeeCount := len(attendees)
	windowStart := day.Add(9 * time.Hour)
	windowEnd := day.Add(17 * time.Hour)
	duration := time.Duration(durationMinutes) * time.Minute
	options := make([]model.ScheduleOption, 0)
	for startsAt := windowStart; !startsAt.Add(duration).After(windowEnd); startsAt = startsAt.Add(15 * time.Minute) {
		endsAt := startsAt.Add(duration)
		attendeesBusy := false
		for _, event := range data.CalendarEvents {
			if _, selected := attendeeSet[event.UserID]; selected && event.StartsAt.Before(endsAt) && event.EndsAt.After(startsAt) {
				attendeesBusy = true
				break
			}
		}
		if attendeesBusy {
			continue
		}

		availableRooms := make([]model.Room, 0)
		for _, room := range rooms {
			if room.Capacity < attendeeCount {
				continue
			}
			busy := false
			for _, booking := range data.Bookings {
				if booking.RoomID == room.ID && booking.StartsAt.Before(endsAt) && booking.EndsAt.After(startsAt) {
					busy = true
					break
				}
			}
			if !busy {
				availableRooms = append(availableRooms, room)
			}
		}
		if len(availableRooms) > 0 {
			options = append(options, model.ScheduleOption{StartsAt: startsAt, EndsAt: endsAt, Rooms: availableRooms})
		}
	}
	return options, nil
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
