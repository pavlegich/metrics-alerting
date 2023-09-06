package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

const table = "metrics"

func NewDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("pgx", path)
	if err != nil {
		return nil, fmt.Errorf("NewDatabase: couldn't open database %w", err)
	}
	defer db.Close()

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
	ctxDB, cancelDB := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelDB()
	if err := db.PingContext(ctxDB); err != nil {
		return fmt.Errorf("SaveToDB: connection to database is died %w", err)
	}

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS $1 (id text NOT NULL, value text NOT NULL);", table)
	if err != nil {
		return fmt.Errorf("SaveToDB: couldn't create table %w", err)
	}

	statement, err := db.Prepare("INSERT INTO $1 (id, value) VALUES ($2, $3) ON CONFLICT (id) DO " +
		"UPDATE SET value=$3 WHERE id=$2")
	if err != nil {
		return fmt.Errorf("SaveToDB: prepare insert statement failed %w", err)
	}
	defer statement.Close()

	for id, value := range DBMetrics {
		if _, err := statement.Exec(table, id, value); err != nil {
			return fmt.Errorf("SaveToDB: statement exec failed %w", err)
		}
	}
	return nil
}
