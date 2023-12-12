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
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/server/handlers"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ps string = "postgresql://localhost:5432/metrics"

func TestStatsStorage_New(t *testing.T) {
	want := &StatStorage{stats: make(map[string]entities.Metrics)}
	assert.Equal(t, want, NewStatStorage(context.Background()))
}

func TestStatStorage_Put(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		stats map[string]entities.Metrics
	}
	type args struct {
		sType  string
		sName  string
		sValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "correct put",
			fields: fields{
				stats: map[string]entities.Metrics{},
			},
			args: args{
				sType:  "gauge",
				sName:  "Gauger",
				sValue: "124.1",
			},
			wantErr: false,
		},
		{
			name: "incorrect gauge",
			fields: fields{
				stats: map[string]entities.Metrics{},
			},
			args: args{
				sType:  "gauge",
				sName:  "Gauger",
				sValue: "value",
			},
			wantErr: true,
		},
		{
			name: "incorrect counter",
			fields: fields{
				stats: map[string]entities.Metrics{},
			},
			args: args{
				sType:  "counter",
				sName:  "Counter",
				sValue: "124.1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tt.fields.stats,
			}
			if err := st.Put(ctx, tt.args.sType, tt.args.sName, tt.args.sValue); (err != nil) != tt.wantErr {
				t.Errorf("StatStorage.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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

func TestStatStorage_SendBatch(t *testing.T) {
	ctx := context.Background()
	ms := storage.NewMemStorage(ctx)
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)
	defer db.Close()
	cfg := &config.ServerConfig{}
	h := handlers.NewWebhook(ctx, ms, db, cfg)
	ts := httptest.NewServer(h.Route(ctx))
	defer ts.Close()
	addr, _ := strings.CutPrefix(ts.URL, "http://")
	gaugeValue := 4.1
	counterValue := int64(3)

	type fields struct {
		stats map[string]entities.Metrics
	}
	type args struct {
		cfg *config.AgentConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				stats: map[string]entities.Metrics{
					"Gauger": {
						ID:    "Gauger",
						MType: "gauge",
						Value: &gaugeValue,
					},
					"Counter": {
						ID:    "Counter",
						MType: "counter",
						Delta: &counterValue,
					},
				},
			},
			args: args{
				cfg: &config.AgentConfig{
					Address: addr,
					Key:     "",
				},
			},
			wantErr: false,
		},
		{
			name: "wrong_url",
			fields: fields{
				stats: map[string]entities.Metrics{
					"Gauger": {
						ID:    "Gauger",
						MType: "gauge",
						Value: &gaugeValue,
					},
					"Counter": {
						ID:    "Counter",
						MType: "counter",
						Delta: &counterValue,
					},
				},
			},
			args: args{
				cfg: &config.AgentConfig{
					Address: "localhost:443",
					Key:     "",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tt.fields.stats,
			}
			if err := st.SendBatch(ctx, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("StatStorage.SendBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatStorage_SendJSON(t *testing.T) {
	// запуск сервера
	ctx := context.Background()
	ms := storage.NewMemStorage(ctx)
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)
	defer db.Close()
	cfg := &config.ServerConfig{}
	h := handlers.NewWebhook(ctx, ms, db, cfg)
	ts := httptest.NewServer(h.Route(ctx))
	defer ts.Close()
	addr, _ := strings.CutPrefix(ts.URL, "http://")
	gaugeValue := 4.1

	type fields struct {
		stats map[string]entities.Metrics
	}
	tests := []struct {
		name   string
		fields fields
		method string
		cfg    *config.AgentConfig
		want   bool
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
			method: http.MethodPost,
			cfg: &config.AgentConfig{
				Address: addr,
				Key:     "",
			},
			want: false,
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
			method: http.MethodPost,
			cfg: &config.AgentConfig{
				Address: "localhost:443",
				Key:     "",
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tc.fields.stats,
			}
			err := st.SendJSON(ctx, tc.cfg)
			if !tc.want {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestStatStorage_SendGZIP(t *testing.T) {
	ctx := context.Background()
	ms := storage.NewMemStorage(ctx)
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)
	defer db.Close()
	cfg := &config.ServerConfig{}
	h := handlers.NewWebhook(ctx, ms, db, cfg)
	ts := httptest.NewServer(h.Route(ctx))
	defer ts.Close()
	addr, _ := strings.CutPrefix(ts.URL, "http://")
	gaugeValue := 4.1
	counterValue := int64(3)

	type fields struct {
		stats map[string]entities.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		cfg     *config.AgentConfig
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				stats: map[string]entities.Metrics{
					"Gauger": {
						ID:    "Gauger",
						MType: "gauge",
						Value: &gaugeValue,
					},
					"Counter": {
						ID:    "Counter",
						MType: "counter",
						Delta: &counterValue,
					},
				},
			},
			cfg: &config.AgentConfig{
				Address: addr,
				Key:     "",
			},
			wantErr: false,
		},
		{
			name: "wrong_url",
			fields: fields{
				stats: map[string]entities.Metrics{
					"Gauger": {
						ID:    "Gauger",
						MType: "gauge",
						Value: &gaugeValue,
					},
					"Counter": {
						ID:    "Counter",
						MType: "counter",
						Delta: &counterValue,
					},
				},
			},
			cfg: &config.AgentConfig{
				Address: "localhost:443",
				Key:     "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tt.fields.stats,
			}
			if err := st.SendGZIP(ctx, tt.cfg); (err != nil) != tt.wantErr {
				t.Errorf("StatStorage.SendBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
