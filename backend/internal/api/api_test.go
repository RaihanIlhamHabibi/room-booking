package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/room-booking/internal/model"
	"example.com/room-booking/internal/repository"
	"example.com/room-booking/internal/service"
	"github.com/labstack/echo/v4"
)

type sessionStore struct {
	repository.Store
	token string
}

func (s *sessionStore) User(_ context.Context, id string) (model.User, error) {
	return model.User{ID: id, Name: "Ada"}, nil
}
func (s *sessionStore) Users(context.Context) ([]model.User, error) {
	return []model.User{{ID: "u-ada", Name: "Ada"}}, nil
}
func (s *sessionStore) CreateSession(_ context.Context, token, _ string, _ time.Time) error {
	s.token = token
	return nil
}
func (s *sessionStore) SessionUser(_ context.Context, token string) (model.User, error) {
	if token != s.token || token == "" {
		return model.User{}, repository.ErrNotFound
	}
	return model.User{ID: "u-ada", Name: "Ada"}, nil
}
func (s *sessionStore) Rooms(context.Context) ([]model.Room, error) { return []model.Room{}, nil }

func TestDemoLoginProtectsRoomRoutes(t *testing.T) {
	store := &sessionStore{}
	e := echo.New()
	New(service.New(store)).Register(e)
	request := func(method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		}
		if cookie != nil {
			req.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		e.ServeHTTP(response, req)
		return response
	}
	if got := request(http.MethodGet, "/rooms", "", nil).Code; got != http.StatusUnauthorized {
		t.Fatalf("anonymous rooms status: %d", got)
	}
	if got := request(http.MethodGet, "/users", "", nil).Code; got != http.StatusOK {
		t.Fatalf("public users status: %d", got)
	}
	login := request(http.MethodPost, "/session", `{"userId":"u-ada"}`, nil)
	if login.Code != http.StatusOK || len(login.Result().Cookies()) != 1 {
		t.Fatalf("login failed: %d", login.Code)
	}
	cookie := login.Result().Cookies()[0]
	if cookie.Name != cookieName || cookie.Path != "/api" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("session cookie options: %+v", cookie)
	}
	if got := request(http.MethodGet, "/rooms", "", cookie).Code; got != http.StatusOK {
		t.Fatalf("signed-in rooms status: %d", got)
	}
}
