package api

import (
	"errors"
	"net/http"
	"time"

	"example.com/room-booking/internal/repository"
	"example.com/room-booking/internal/service"
	"github.com/labstack/echo/v4"
)

const cookieName = "room_session"

type Handler struct{ service *service.Service }

func New(s *service.Service) *Handler { return &Handler{service: s} }

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"status": "ok"}) })
	e.GET("/users", h.users)
	e.POST("/session", h.login)
	authorized := e.Group("", h.requireSession)
	authorized.GET("/session", h.session)
	authorized.DELETE("/session", h.logout)
	authorized.GET("/rooms", h.rooms)
	authorized.GET("/rooms/:id", h.room)
	authorized.PATCH("/rooms/:id", h.updateRoom)
	authorized.GET("/rooms/:id/bookings", h.bookings)
	authorized.GET("/calendar-events", h.calendarEvents)
}

func respondError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, service.ErrInvalid):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input"})
	default:
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

func token(c echo.Context) string {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (h *Handler) requireSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, err := h.service.SessionUser(c.Request().Context(), token(c)); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "sign in required"})
		}
		return next(c)
	}
}

func (h *Handler) users(c echo.Context) error {
	users, err := h.service.Users(c.Request().Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": users})
}

func (h *Handler) login(c echo.Context) error {
	var input struct {
		UserID string `json:"userId"`
	}
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}
	user, secret, err := h.service.Login(c.Request().Context(), input.UserID)
	if err != nil {
		return respondError(c, err)
	}
	c.SetCookie(&http.Cookie{Name: cookieName, Value: secret, Path: "/api", HttpOnly: true, SameSite: http.SameSiteLaxMode, Expires: time.Now().Add(7 * 24 * time.Hour)})
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) session(c echo.Context) error {
	user, err := h.service.SessionUser(c.Request().Context(), token(c))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "sign in required"})
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) logout(c echo.Context) error {
	if err := h.service.Logout(c.Request().Context(), token(c)); err != nil {
		return respondError(c, err)
	}
	c.SetCookie(&http.Cookie{Name: cookieName, Value: "", Path: "/api", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) rooms(c echo.Context) error {
	rooms, err := h.service.Rooms(c.Request().Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": rooms})
}

func (h *Handler) room(c echo.Context) error {
	room, err := h.service.Room(c.Request().Context(), c.Param("id"))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, room)
}

func (h *Handler) updateRoom(c echo.Context) error {
	var input struct {
		Description string `json:"description"`
	}
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
	}
	room, err := h.service.UpdateRoomDescription(c.Request().Context(), c.Param("id"), input.Description)
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, room)
}

func (h *Handler) bookings(c echo.Context) error {
	rows, err := h.service.Bookings(c.Request().Context(), c.Param("id"))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": rows})
}

func (h *Handler) calendarEvents(c echo.Context) error {
	rows, err := h.service.CalendarEvents(c.Request().Context(), c.QueryParam("date"))
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"data": rows})
}
