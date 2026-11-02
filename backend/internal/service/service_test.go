package service

import (
	"context"
	"errors"
	"strings"
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
	booking      model.Booking
	bookingErr   error
	available    []model.Room
	availableAt  time.Time
	availableEnd time.Time
	attendees    int
	users        []model.User
	rooms        []model.Room
	scheduleData repository.ScheduleData
	scheduleDay  time.Time
	scheduleIDs  []string
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
func (f *fakeStore) Room(_ context.Context, id string) (model.Room, error) {
	return model.Room{ID: id}, nil
}
func (f *fakeStore) CreateBooking(_ context.Context, booking model.Booking) error {
	f.booking = booking
	return f.bookingErr
}
func (f *fakeStore) AvailableRooms(_ context.Context, startsAt, endsAt time.Time, attendeeCount int) ([]model.Room, error) {
	f.availableAt, f.availableEnd, f.attendees = startsAt, endsAt, attendeeCount
	return f.available, nil
}
func (f *fakeStore) Users(context.Context) ([]model.User, error) {
	if f.users != nil {
		return f.users, nil
	}
	return []model.User{{ID: "u-ada"}, {ID: "u-bima"}, {ID: "u-citra"}}, nil
}
func (f *fakeStore) Rooms(context.Context) ([]model.Room, error) {
	return f.rooms, nil
}
func (f *fakeStore) ScheduleData(_ context.Context, day time.Time, attendeeIDs []string) (repository.ScheduleData, error) {
	f.scheduleDay = day
	f.scheduleIDs = attendeeIDs
	return f.scheduleData, nil
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

func TestCreateBookingTrimsPurposeAndUsesCurrentOrganizer(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)
	booking, err := svc.CreateBooking(context.Background(), "u-ada", "r-garden", "2030-05-13", "09:30", "10:30", "  Study group  ")
	if err != nil {
		t.Fatal(err)
	}
	if booking.ID == "" || booking.ID != store.booking.ID || booking.OrganizerID != "u-ada" || booking.Purpose != "Study group" {
		t.Fatalf("unexpected booking: %+v", booking)
	}
	if booking.StartsAt.Format("-07:00") != "+07:00" || booking.StartsAt.Hour() != 9 || booking.EndsAt.Hour() != 10 {
		t.Fatalf("booking was not parsed in Asia/Jakarta: %+v", booking)
	}
}

func TestCreateBookingRejectsInvalidInputBeforePersistence(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)
	for _, input := range []struct {
		date, start, end, purpose string
	}{
		{"2030-02-30", "09:00", "10:00", "Meeting"},
		{"2030-05-13", "10:00", "09:00", "Meeting"},
		{"2030-05-13", "09:00", "10:00", "   "},
		{"2030-05-13", "09:00", "10:00", strings.Repeat("x", 201)},
	} {
		if _, err := svc.CreateBooking(context.Background(), "u-ada", "r-garden", input.date, input.start, input.end, input.purpose); !errors.Is(err, ErrInvalid) {
			t.Fatalf("input %+v: %v", input, err)
		}
	}
	if store.booking.ID != "" {
		t.Fatal("invalid booking reached persistence")
	}
}

func TestAvailableRoomsValidatesAndPassesTimeAndAttendeeCount(t *testing.T) {
	store := &fakeStore{available: []model.Room{{ID: "r-maple"}}}
	svc := New(store)
	rooms, err := svc.AvailableRooms(context.Background(), "2030-05-13", "10:00", "11:30", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 1 || rooms[0].ID != "r-maple" || store.attendees != 5 {
		t.Fatalf("unexpected availability result: %+v, attendees=%d", rooms, store.attendees)
	}
	if store.availableAt.Hour() != 10 || store.availableAt.Format("-07:00") != "+07:00" || store.availableEnd.Hour() != 11 || store.availableEnd.Minute() != 30 {
		t.Fatalf("unexpected search interval: %v - %v", store.availableAt, store.availableEnd)
	}
	if _, err := svc.AvailableRooms(context.Background(), "2030-05-13", "10:00", "11:30", 0); !errors.Is(err, ErrInvalid) {
		t.Fatalf("zero attendees: %v", err)
	}
}

func TestCreateBookingReturnsConflictFromRepository(t *testing.T) {
	store := &fakeStore{bookingErr: repository.ErrConflict}
	_, err := New(store).CreateBooking(context.Background(), "u-ada", "r-garden", "2030-05-13", "09:00", "10:00", "Meeting")
	if !errors.Is(err, repository.ErrConflict) {
		t.Fatalf("booking conflict: %v", err)
	}
}

func TestScheduleFindsSlotsSharedByOrganizerAndAttendees(t *testing.T) {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStore{
		rooms: []model.Room{{ID: "r-tiny", Capacity: 1}, {ID: "r-small", Capacity: 2}, {ID: "r-large", Capacity: 5}},
		scheduleData: repository.ScheduleData{
			CalendarEvents: []model.CalendarEvent{
				{UserID: "u-ada", StartsAt: time.Date(2030, 5, 13, 9, 0, 0, 0, location), EndsAt: time.Date(2030, 5, 13, 10, 0, 0, 0, location)},
				{UserID: "u-bima", StartsAt: time.Date(2030, 5, 13, 10, 0, 0, 0, location), EndsAt: time.Date(2030, 5, 13, 10, 30, 0, 0, location)},
				{UserID: "u-citra", StartsAt: time.Date(2030, 5, 13, 9, 0, 0, 0, location), EndsAt: time.Date(2030, 5, 13, 17, 0, 0, 0, location)},
			},
			Bookings: []model.Booking{
				{RoomID: "r-small", StartsAt: time.Date(2030, 5, 13, 10, 30, 0, 0, location), EndsAt: time.Date(2030, 5, 13, 11, 0, 0, 0, location)},
				{RoomID: "r-large", StartsAt: time.Date(2030, 5, 13, 11, 0, 0, 0, location), EndsAt: time.Date(2030, 5, 13, 11, 30, 0, 0, location)},
			},
		},
	}

	options, err := New(store).Schedule(context.Background(), "u-ada", "2030-05-13", 30, []string{"u-ada", "u-bima"})
	if err != nil {
		t.Fatal(err)
	}
	if len(options) == 0 || options[0].StartsAt.Hour() != 10 || options[0].StartsAt.Minute() != 30 {
		t.Fatalf("first shared free slot should be 10:30, got %+v", options)
	}
	if len(options[0].Rooms) != 1 || options[0].Rooms[0].ID != "r-large" || options[0].Rooms[0].Capacity < 2 {
		t.Fatalf("slot did not exclude the busy room or fit both people: %+v", options[0].Rooms)
	}
	for _, option := range options {
		for _, room := range option.Rooms {
			if room.ID == "r-tiny" {
				t.Fatalf("returned a room smaller than the attendee group: %+v", option)
			}
		}
	}
	for _, option := range options {
		if option.StartsAt.Hour() == 9 || option.StartsAt.Hour() == 10 {
			if option.StartsAt.Hour() != 10 || option.StartsAt.Minute() != 30 {
				t.Fatalf("returned a slot overlapping an attendee event: %+v", option)
			}
		}
	}
	if store.scheduleDay.Hour() != 0 || store.scheduleDay.Format("-07:00") != "+07:00" || len(store.scheduleIDs) != 2 {
		t.Fatalf("wrong calendar query: day=%v ids=%v", store.scheduleDay, store.scheduleIDs)
	}
}

func TestScheduleRejectsInvalidAttendeesDateAndDuration(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)
	for _, input := range []struct {
		date     string
		duration int
		people   []string
	}{
		{"2030-02-30", 30, []string{"u-bima"}},
		{"2030-05-13", 20, []string{"u-bima"}},
		{"2030-05-13", 30, []string{"unknown"}},
		{"2030-05-13", 30, nil},
	} {
		if _, err := svc.Schedule(context.Background(), "u-ada", input.date, input.duration, input.people); !errors.Is(err, ErrInvalid) {
			t.Fatalf("input %+v: %v", input, err)
		}
	}
	if len(store.scheduleIDs) != 0 {
		t.Fatal("invalid scheduling input reached schedule query")
	}
}
