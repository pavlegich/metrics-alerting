package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

// DBMetric содержит название и значение метрики
// для хранения в базе данных.
type DBMetric struct {
	ID    string
	Value string
}

// Database содержит информацию о базе данных.
type Database struct {
	db *sql.DB
}

// NewDatabase создаёт новый объект Database для хранения метрик сервера.
func NewDatabase(db *sql.DB) *Database {
	return &Database{
		db: db,
	}
}

// Save сохраняет все метрики из хранилища сервера в базу данных.
func (d *Database) Save(ctx context.Context, ms interfaces.MetricStorage) error {
	// Получение всех метрик из хранилища
	metrics := ms.GetAll(ctx)
	DBMetrics := make(map[string]string)
	for m, v := range metrics {
		DBMetrics[m] = v
	}

	// Проверка базы данных
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("SaveToDB: connection to database is died %w", err)
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("SaveToDB: begin transaction failed %w", err)
	}
	defer tx.Rollback()

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

// Load получает все метрики из хранилища
// и сохраняет их в хранилище сервера.
func (d *Database) Load(ctx context.Context, ms interfaces.MetricStorage) error {
	// Проверка базы данных
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("LoadFromDB: connection to database is died %w", err)
	}

	// Получение метрик из хранилища
	rows, err := d.db.QueryContext(ctx, "SELECT id, value FROM storage")
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

	// Сохранение данных в локальном хранилище
	for _, metric := range DBMetrics {
		// Пока все будут gauge, чтобы ошибок с конвертацией не было, тип не хранится в MemStorage
		if status := ms.Put(ctx, "gauge", metric.ID, metric.Value); status != http.StatusOK {
			return fmt.Errorf("LoadFromDB: put all metrics status %v", status)
		}
	}

	return nil
}

// Ping проверяет наличие соединения с базой данных.
func (d *Database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}
