package interfaces

import "context"

// Server содержит методы для работы сервера
type Server interface {
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context) error
	GetAddress(ctx context.Context) string
}
