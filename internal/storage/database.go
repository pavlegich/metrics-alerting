package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

type DBMetric struct {
	ID    string
	Value string
}

func NewDatabase(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("pgx", path)
	if err != nil {
		return nil, fmt.Errorf("NewDatabase: couldn't open database %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("NewDatabase: connection with database is died %w", err)
	}

	return db, nil
}

func SaveToDB(ctx context.Context, db *sql.DB, ms interfaces.MetricStorage) error {
	// Получение всех метрик из хранилища
	metrics, status := ms.GetAll(ctx)
	if status != http.StatusOK {
		return fmt.Errorf("SaveToDB: metrics get error %v", status)
	}
	DBMetrics := make(map[string]string)
	for m, v := range metrics {
		DBMetrics[m] = v
	}

	// Проверка базы данных
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("SaveToDB: connection to database is died %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("SaveToDB: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Создание таблицы при её отсутствии
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS storage (
		id text PRIMARY KEY, 
		value text NOT NULL);`); err != nil {
		return fmt.Errorf("SaveToDB: couldn't create table %w", err)
	}

	// Сохранение метрик в хранилище
	statement, err := tx.PrepareContext(ctx, "INSERT INTO storage (id, value) VALUES ($1, $2) ON CONFLICT (id) DO "+
		"UPDATE SET value=$2 WHERE storage.id=$1")
	if err != nil {
		return fmt.Errorf("SaveToDB: insert into table failed %w", err)
	}
	defer statement.Close()

	for id, value := range DBMetrics {
		if _, err := statement.ExecContext(ctx, id, value); err != nil {
			return fmt.Errorf("SaveToDB: statement exec failed %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("SaveToDB: commit transaction failed %w", err)
	}

	return nil
}

func LoadFromDB(ctx context.Context, db *sql.DB, ms interfaces.MetricStorage) error {
	// Проверка базы данных
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("LoadFromDB: connection to database is died %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("LoadFromDB: begin transaction failed %w", err)
	}
	defer tx.Rollback()

	// Создание таблицы при её отсутствии
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS storage (
		id text PRIMARY KEY, 
		value text NOT NULL);`); err != nil {
		return fmt.Errorf("LoadFromDB: couldn't create table %w", err)
	}

	// Получение метрик из хранилища
	rows, err := tx.QueryContext(ctx, "SELECT id, value FROM storage")
	if err != nil {
		return fmt.Errorf("LoadFromDB: read rows from table failed %w", err)
	}
	defer rows.Close()

	DBMetrics := make([]DBMetric, 0)
	for rows.Next() {
		var metric DBMetric
		err = rows.Scan(&metric.ID, &metric.Value)
		if err != nil {
			return fmt.Errorf("LoadFromDB: scan row failed %w", err)
		}
		DBMetrics = append(DBMetrics, metric)
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("LoadFromDB: rows.Err %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("LoadFromDB: commit transaction failed %w", err)
	}

	// Сохранение данных в локальном хранилище
	for _, metric := range DBMetrics {
		// Пока все будут gauge, чтобы ошибок с конвертацией не было, тип не хранится в MemStorage
		if status := ms.Put(ctx, "gauge", metric.ID, metric.Value); status != http.StatusOK {
			return fmt.Errorf("LoadFromDB: put all metrics status %v", status)
		}
	}

	return nil
}
