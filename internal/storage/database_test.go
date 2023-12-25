package storage

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/stretchr/testify/require"
)

var ps string = "postgresql://localhost:5432/metrics"

func TestSaveToDB(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)

	type args struct {
		db *sql.DB
		ms interfaces.MetricStorage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				db: db,
				ms: &MemStorage{
					Metrics: map[string]string{
						"Gauger":  "241.4",
						"Counter": "4",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := NewDatabase(tt.args.db)
			if err := database.Save(ctx, tt.args.ms); (err != nil) != tt.wantErr {
				t.Errorf("SaveToDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFromDB(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)

	type args struct {
		db *sql.DB
		ms interfaces.MetricStorage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				db: db,
				ms: &MemStorage{
					Metrics: map[string]string{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := NewDatabase(tt.args.db)
			if err := database.Load(ctx, tt.args.ms); (err != nil) != tt.wantErr {
				t.Errorf("LoadFromDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
