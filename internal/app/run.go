package app

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/logger"
	"github.com/pavlegich/metrics-alerting/internal/middlewares"
	"github.com/pavlegich/metrics-alerting/internal/models"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"go.uber.org/zap"
)

var (
	StoreInterval time.Duration
	StoragePath   *string
	Restore       *bool
)

// функция run запускает сервер
func Run() error {
	// Считывание флага адреса и его запись в структуру
	addr := models.NewAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "HTTP-server endpoint address host:port")

	store := flag.Int("i", 300, "Frequency of storing on disk")
	StoragePath = flag.String("f", "/tmp/metrics-db.json", "Full path of values storage")
	Restore = flag.Bool("r", true, "Restore values from the disk")

	flag.Parse()

	// Проверяем переменные окружения
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr.Set(envAddr)
	}
	if envStore := os.Getenv("STORE_INTERVAL"); envStore != "" {
		var err error
		*store, err = strconv.Atoi(envStore)
		if err != nil {
			return err
		}
	}
	if envStoragePath := os.Getenv("FILE_STORAGE_PATH"); envStoragePath != "" {
		*StoragePath = envStoragePath
	}
	if envStore := os.Getenv("RESTORE"); envStore != "" {
		var err error
		*Restore, err = strconv.ParseBool(envStore)
		if err != nil {
			return err
		}
	}

	if *store == 0 {
		StoreInterval = time.Duration(1) * time.Second
	} else {
		StoreInterval = time.Duration(*store) * time.Second
	}

	if err := logger.Initialize("Info"); err != nil {
		return err
	}
	defer logger.Log.Sync()

	// Создание хранилища метрик
	memStorage := storage.NewMemStorage()

	// Загрузка данных из файла
	if *Restore {
		var err error
		memStorage, err = storage.LoadStorage(*StoragePath)
		if err != nil {
			return err
		}
	}

	// Создание нового хендлера для сервера
	webhook := handlers.NewWebhook(memStorage)

	if *StoragePath != "" {
		go storeMetricsRoutine(webhook, StoreInterval, *StoragePath)
	}

	r := chi.NewRouter()
	r.Use(middlewares.Recovery)
	r.Mount("/", webhook.Route())

	logger.Log.Info("Running server", zap.String("address", addr.String()))

	return http.ListenAndServe(addr.String(), r)
}

func storeMetricsRoutine(wh *handlers.Webhook, store time.Duration, path string) error {
	for {
		if err := storage.SaveStorage(path, &wh.MemStorage); err != nil {
			return err
		}
		time.Sleep(store)
	}
}
