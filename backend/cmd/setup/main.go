package main

import (
	"context"
	"log"
	"os"
	"time"

	"example.com/room-booking/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is required")
	}
	config, err := pgx.ParseConfig(url)
	if err != nil {
		log.Fatal(err)
	}
	// Embedded migrations contain multiple SQL statements in each version.
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	db := stdlib.OpenDB(*config)
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := database.Setup(ctx, db); err != nil {
		log.Fatal(err)
	}
	log.Print("database ready")
}
