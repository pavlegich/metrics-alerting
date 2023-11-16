package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Init(ctx context.Context, path string) (*sql.DB, error) {
	// Открытие и проверка базы данных
	db, err := sql.Open("pgx", path)
	if err != nil {
		return nil, fmt.Errorf("Init: couldn't open database %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("Init: connection with database is died %w", err)
	}

	// Миграции
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("Init: goose set dialect failed %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("Init: goose up failed %w", err)
	}

	return db, nil
}
