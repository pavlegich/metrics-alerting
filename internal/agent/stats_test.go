package agent

import (
	"runtime"
	"testing"

	"github.com/pavlegich/metrics-alerting/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestStatStorage_Update(t *testing.T) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	type fields struct {
		stats map[string]models.Metrics
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
				stats: map[string]models.Metrics{},
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
				stats: map[string]models.Metrics{},
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
	want := &StatStorage{stats: make(map[string]models.Metrics)}
	assert.Equal(t, want, NewStatStorage())
}

// func TestMemStorage_Send(t *testing.T) {
// 	// запуск сервера
// 	ms := storage.NewMemStorage()
// 	h := handlers.NewWebhook(ms)
// 	ts := httptest.NewServer(h.Route())
// 	defer ts.Close()
// 	addr, _ := strings.CutPrefix(ts.URL, "http://")
// 	gaugeValue := float64(4.1)
// 	counterValue := int64(4)

// 	type fields struct {
// 		stats map[string]models.Metrics
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		method  string
// 		address string
// 		want    bool
// 	}{
// 		{
// 			name: "successful_gauge_request",
// 			fields: fields{
// 				stats: map[string]models.Metrics{
// 					"SomeMetric": {
// 						ID:    "SomeMetric",
// 						MType: "gauge",
// 						Value: &gaugeValue,
// 					},
// 				},
// 			},
// 			method:  http.MethodPost,
// 			address: addr,
// 			want:    false,
// 		},
// 		{
// 			name: "successful_counter_request",
// 			fields: fields{
// 				stats: map[string]models.Metrics{
// 					"SomeMetric": {
// 						ID:    "SomeMetric",
// 						MType: "counter",
// 						Delta: &counterValue,
// 					},
// 				},
// 			},
// 			method:  http.MethodPost,
// 			address: addr,
// 			want:    false,
// 		},
// 		{
// 			name: "wrong_address",
// 			fields: fields{
// 				stats: map[string]models.Metrics{
// 					"SomeMetric": {
// 						ID:    "SomeMetric",
// 						MType: "gauge",
// 						Value: &gaugeValue,
// 					},
// 				},
// 			},
// 			method:  http.MethodPost,
// 			address: "localhost:443",
// 			want:    true,
// 		},
// 	}
// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			st := &StatStorage{
// 				stats: tc.fields.stats,
// 			}
// 			err := st.Send(tc.address)
// 			if !tc.want {
// 				assert.NoError(t, err)
// 				return
// 			}
// 			assert.Error(t, err)
// 		})
// 	}
// }
