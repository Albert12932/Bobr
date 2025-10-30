package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB() (*pgxpool.Pool, error) {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user, password, host, port, dbname,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Ошибка создания пула подключений: %v", err)
	}
	defer cancel()
	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	cancel()

	return pool, nil
}
