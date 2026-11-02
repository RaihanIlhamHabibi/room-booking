package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/room-booking/internal/api"
	"example.com/room-booking/internal/repository"
	"example.com/room-booking/internal/service"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openDB() (*gorm.DB, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is required")
	}
	return gorm.Open(postgres.Open(url), &gorm.Config{})
}

func newEcho(handler *api.Handler) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	handler.Register(e)
	return e
}

func run(lifecycle fx.Lifecycle, e *echo.Echo) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: ":" + port, Handler: e}
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Print(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return server.Shutdown(shutdownCtx)
		},
	})
}

func main() {
	fx.New(fx.Provide(openDB, repository.New, service.New, api.New, newEcho), fx.Invoke(run)).Run()
}
