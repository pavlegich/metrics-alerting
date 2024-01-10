package interfaces

import "context"

type Server interface {
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context) error
	GetAddress(ctx context.Context) string
}
