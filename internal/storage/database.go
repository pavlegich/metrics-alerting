package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

func NewDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("pgx", path)
	if err != nil {
		return nil, fmt.Errorf("NewDatabase: couldn't open database %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("NewDatabase: connection with database is died %w", err)
	}

	return db, nil
}

func SaveToDB(db *sql.DB, ms interfaces.MetricStorage) error {
	// Получение всех метрик из хранилища
	metrics, status := ms.GetAll()
	if status != http.StatusOK {
		return fmt.Errorf("SaveToDB: metrics get error %v", status)
	}
	DBMetrics := make(map[string]string)
	for m, v := range metrics {
		DBMetrics[m] = v
	}

	// Проверка базы данных
	if err := db.Ping(); err != nil {
		return fmt.Errorf("SaveToDB: connection to database is died %w", err)
	}

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS storage (id text PRIMARY KEY, value text NOT NULL);")
	if err != nil {
		return fmt.Errorf("SaveToDB: couldn't create table %w", err)
	}

	statement, err := db.Prepare("INSERT INTO storage (id, value) VALUES ($1, $2) ON CONFLICT (id) DO " +
		"UPDATE SET value=$2 WHERE storage.id=$1")
	if err != nil {
		return fmt.Errorf("SaveToDB: prepare insert statement failed %w", err)
	}
	defer statement.Close()

	for id, value := range DBMetrics {
		if _, err := statement.Exec(id, value); err != nil {
			return fmt.Errorf("SaveToDB: statement exec failed %w", err)
		}
	}
	return nil
}
