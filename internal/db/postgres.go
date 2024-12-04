package db

import (
	"context"
	"fmt"
	"log"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/brianvoe/gofakeit"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Database interface {
	SaveRefreshToken(ctx context.Context, userID string, refreshTokenHash string, ipAddress string) error
	GetRefreshTokenHash(ctx context.Context, userID string) (string, error)
	UpdateRefreshToken(ctx context.Context, userID string, newRefreshTokenHash string) error
}

type DB struct {
	Pool *pgxpool.Pool
}

func InitDB(ctx context.Context) (*DB, error) {
    var err error
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        return nil, fmt.Errorf("переменная окружения DATABASE_URL не установлена")
    }

    pool, err := pgxpool.New(ctx, dbURL)
    if err != nil {
        return nil, fmt.Errorf("ошибка при подключении к базе данных: %v", err)
    }

	// Проверяем соединение
    conn, err := pool.Acquire(ctx)
    if err != nil {
        log.Fatalf("Ошибка при проверке соединения с базой данных: %v", err)
        return nil, fmt.Errorf("Ошибка при проверке соединения с базой данных: %v", err)
    }
    conn.Release()

    return &DB{Pool: pool}, nil
}

// Сохранение Refresh токена
func (db *DB) SaveRefreshToken(ctx context.Context, userID string, refreshTokenHash string, ipAddress string) error {
    builderInsert := sq.Insert("users").
        PlaceholderFormat(sq.Dollar).
        Columns("user_id","email", "refresh_token_hash", "last_ip_address").
        Values(userID, gofakeit.Email(), refreshTokenHash, ipAddress).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET " +
			"refresh_token_hash = EXCLUDED.refresh_token_hash, " +
			"last_ip_address = EXCLUDED.last_ip_address")

    query, args, err := builderInsert.ToSql()
    if err != nil {
        return fmt.Errorf("ошибка при построении запроса: %v", err)
    }

    // Выполняем запрос в базе данных
    _, err = db.Pool.Exec(ctx, query, args...)
    if err != nil {
        return fmt.Errorf("ошибка при выполнении запроса: %v", err)
    }

    return nil
}

// Получение хеша Refresh токена
func (db *DB) GetRefreshTokenHash(ctx context.Context, userID string) (string, error) {
    builderSelect := sq.Select("refresh_token_hash").
        From("users").
        Where(sq.Eq{"user_id": userID}).
        PlaceholderFormat(sq.Dollar)

    // Генерируем SQL-запрос и аргументы
    query, args, err := builderSelect.ToSql()
    if err != nil {
        return "", fmt.Errorf("ошибка при построении запроса: %v", err)
    }

    var refreshTokenHash string

    // Выполняем запрос в базе данных
    err = db.Pool.QueryRow(ctx, query, args...).Scan(&refreshTokenHash)
    if err != nil {
        return "", fmt.Errorf("ошибка при выполнении запроса: %v", err)
    }

    return refreshTokenHash, nil
}

// Обновление хеша refresh токенов
func (db *DB) UpdateRefreshToken(ctx context.Context, userID string, newRefreshTokenHash string) error {

    builderUpdate := sq.Update("users").
        Set("refresh_token_hash", newRefreshTokenHash).
        Where(sq.Eq{"user_id": userID}).
        PlaceholderFormat(sq.Dollar)

    query, args, err := builderUpdate.ToSql()
    if err != nil {
        return fmt.Errorf("ошибка при построении запроса: %v", err)
    }

    _, err = db.Pool.Exec(ctx, query, args...)
    if err != nil {
        return fmt.Errorf("ошибка при выполнении запроса: %v", err)
    }

    return nil
}