package agent

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/server/handlers"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ps string = "postgresql://localhost:5432/metrics"

func TestStatStorage_Update(t *testing.T) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	type fields struct {
		stats map[string]entities.Metrics
	}
	type args struct {
		memStats runtime.MemStats
		count    int
		rand     float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "update_stat",
			fields: fields{
				stats: map[string]entities.Metrics{},
			},
			args: args{
				memStats: ms,
				count:    5,
				rand:     83.2,
			},
			want: false,
		},
		{
			name: "update_stat",
			fields: fields{
				stats: map[string]entities.Metrics{},
			},
			args: args{
				memStats: ms,
				count:    5,
				rand:     83.2,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tt.fields.stats,
			}
			err := st.Update(context.Background(), tt.args.memStats, tt.args.count, tt.args.rand)
			if !tt.want {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestStatsStorage_New(t *testing.T) {
	want := &StatStorage{stats: make(map[string]entities.Metrics)}
	assert.Equal(t, want, NewStatStorage(context.Background()))
}

func TestMemStorage_Send(t *testing.T) {
	// запуск сервера
	ctx := context.Background()
	ms := storage.NewMemStorage(ctx)
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)
	defer db.Close()
	h := handlers.NewWebhook(ctx, ms, db)
	ts := httptest.NewServer(h.Route(ctx))
	defer ts.Close()
	addr, _ := strings.CutPrefix(ts.URL, "http://")
	gaugeValue := float64(4.1)
	counterValue := int64(4)

	type fields struct {
		stats map[string]entities.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		method  string
		address string
		key     string
		want    bool
	}{
		{
			name: "successful_gauge_request",
			fields: fields{
				stats: map[string]entities.Metrics{
					"SomeMetric": {
						ID:    "SomeMetric",
						MType: "gauge",
						Value: &gaugeValue,
					},
				},
			},
			method:  http.MethodPost,
			address: addr,
			key:     "",
			want:    false,
		},
		{
			name: "successful_counter_request",
			fields: fields{
				stats: map[string]entities.Metrics{
					"SomeMetric": {
						ID:    "SomeMetric",
						MType: "counter",
						Delta: &counterValue,
					},
				},
			},
			method:  http.MethodPost,
			address: addr,
			key:     "",
			want:    false,
		},
		{
			name: "wrong_address",
			fields: fields{
				stats: map[string]entities.Metrics{
					"SomeMetric": {
						ID:    "SomeMetric",
						MType: "gauge",
						Value: &gaugeValue,
					},
				},
			},
			method:  http.MethodPost,
			address: "localhost:443",
			key:     "",
			want:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tc.fields.stats,
			}
			err := st.SendJSON(ctx, tc.address, tc.key)
			if !tc.want {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}
