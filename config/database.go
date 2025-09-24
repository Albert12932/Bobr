package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func ConnectDB() (*pgxpool.Pool, error) {
	// Загружаем .env (если есть)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Файл .env не найден, используем переменные окружения из системы")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Проверяем, что все переменные заданы
	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("не заданы переменные окружения для подключения к БД: DB_HOST=%q, DB_PORT=%q, DB_USER=%q, DB_NAME=%q",
			host, port, user, dbname)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname,
	)

	// Повторяем попытки подключения до 15 раз с интервалом 1 секунда
	maxRetries := 15
	for i := 1; i <= maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pool, err := pgxpool.New(ctx, dsn)
		cancel() // освобождаем контекст сразу

		if err != nil {
			log.Printf("🔁 Попытка %d/%d: ошибка создания пула: %v", i, maxRetries, err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Проверяем подключение
		ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
		err = pool.Ping(ctx)
		cancel()

		if err == nil {
			log.Println("✅ Успешное подключение к базе данных")
			return pool, nil
		}

		log.Printf("🔁 Попытка %d/%d: ping не удался: %v", i, maxRetries, err)
		pool.Close() // закрываем пул перед повторной попыткой
		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("не удалось подключиться к БД после %d попыток", maxRetries)
}
