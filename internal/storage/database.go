package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

type DBMetric struct {
	Id    string
	Value string
}

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

	// Создание таблицы при её отсутствии
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS storage (id text PRIMARY KEY, value text NOT NULL);"); err != nil {
		return fmt.Errorf("SaveToDB: couldn't create table %w", err)
	}

	// Сохранение метрик в хранилище
	statement, err := db.Prepare("INSERT INTO storage (id, value) VALUES ($1, $2) ON CONFLICT (id) DO " +
		"UPDATE SET value=$2 WHERE storage.id=$1")
	if err != nil {
		return fmt.Errorf("SaveToDB: insert into table failed %w", err)
	}
	defer statement.Close()

	for id, value := range DBMetrics {
		if _, err := statement.Exec(id, value); err != nil {
			return fmt.Errorf("SaveToDB: statement exec failed %w", err)
		}
	}

	return nil
}

func LoadFromDB(db *sql.DB, ms interfaces.MetricStorage) error {
	// Проверка базы данных
	if err := db.Ping(); err != nil {
		return fmt.Errorf("SaveToDB: connection to database is died %w", err)
	}

	// Создание таблицы при её отсутствии
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS storage (id text PRIMARY KEY, value text NOT NULL);"); err != nil {
		return fmt.Errorf("LoadFromDB: couldn't create table %w", err)
	}

	// Получение метрик из хранилища
	rows, err := db.Query("SELECT id, value FROM storage")
	if err != nil {
		return fmt.Errorf("LoadFromDB: read rows from table failed %w", err)
	}
	defer rows.Close()

	DBMetrics := make([]DBMetric, 0)
	for rows.Next() {
		var metric DBMetric
		err = rows.Scan(&metric.Id, &metric.Value)
		if err != nil {
			return fmt.Errorf("LoadFromDB: scan row failed %w", err)
		}
		DBMetrics = append(DBMetrics, metric)
	}

	for _, metric := range DBMetrics {
		// Сейчас все пусть будут gauge, чтобы ошибок с конвертацией не было, он не записывает тип в storage
		// Впоследствии сделаю, чтобы в storage хранились отдельно gauge и counter, не все string
		if status := ms.Put("gauge", metric.Id, metric.Value); status != http.StatusOK {
			return fmt.Errorf("LoadFromDB: get all metrics status %v", status)
		}
	}

	return nil
}
