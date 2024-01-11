package storage

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pavlegich/metrics-alerting/internal/mocks"
)

func TestDatabase_Save(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mocks.NewMockStorage(ctrl)

	gomock.InOrder(
		mock.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil),
	)

	type args struct {
		metrics map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				metrics: map[string]string{
					"Gauger":  "241.4",
					"Counter": "4",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Metrics: tt.args.metrics,
			}
			if err := mock.Save(ctx, ms); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDatabase_Load(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mocks.NewMockStorage(ctrl)

	gomock.InOrder(
		mock.EXPECT().Load(gomock.Any(), gomock.Any()).Return(nil),
	)

	type args struct {
		metrics map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				metrics: map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Metrics: tt.args.metrics,
			}
			if err := mock.Load(ctx, ms); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
