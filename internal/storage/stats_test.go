package storage

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestStatStorage_Update(t *testing.T) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	type fields struct {
		stats map[string]models.Stat
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
				stats: map[string]models.Stat{},
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
				stats: map[string]models.Stat{},
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
			err := st.Update(tt.args.memStats, tt.args.count, tt.args.rand)
			if !tt.want {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestStatsStorage_New(t *testing.T) {
	want := &StatStorage{stats: make(map[string]models.Stat)}
	assert.Equal(t, want, NewStatStorage())
}

func TestMemStorage_Send(t *testing.T) {
	// запуск сервера
	ms := NewMemStorage()
	h := handlers.NewWebhook(ms)
	ts := httptest.NewServer(h.Route())
	defer ts.Close()
	addr, _ := strings.CutPrefix(ts.URL, "http://")

	type fields struct {
		stats map[string]models.Stat
	}
	tests := []struct {
		name    string
		fields  fields
		method  string
		address string
		want    bool
	}{
		{
			name: "successful_request",
			fields: fields{
				stats: map[string]models.Stat{
					"SomeMetric": {
						Type:  "gauge",
						Name:  "SomeMetric",
						Value: "4.1",
					},
				},
			},
			method:  http.MethodPost,
			address: addr,
			want:    false,
		},
		{
			name: "wrong_address",
			fields: fields{
				stats: map[string]models.Stat{
					"SomeMetric": {
						Type:  "gauge",
						Name:  "SomeMetric",
						Value: "4.1",
					},
				},
			},
			method:  http.MethodPost,
			address: "localhost:443",
			want:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tc.fields.stats,
			}
			err := st.Send(tc.address)
			if !tc.want {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}
