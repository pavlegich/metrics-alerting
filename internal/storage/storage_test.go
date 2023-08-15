package storage

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_Put(t *testing.T) {
	type fields struct {
		Metrics map[string]string
	}
	type args struct {
		metricType  string
		metricName  string
		metricValue string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "put_new_gauge",
			fields: fields{
				Metrics: map[string]string{},
			},
			args: args{
				metricType:  "gauge",
				metricName:  "SomeMetric",
				metricValue: "844082.1",
			},
			want: http.StatusOK,
		},
		{
			name: "put_wrong_gauge",
			fields: fields{
				Metrics: map[string]string{},
			},
			args: args{
				metricType:  "gauge",
				metricName:  "SomeMetric",
				metricValue: "none",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "put_new_counter",
			fields: fields{
				Metrics: map[string]string{},
			},
			args: args{
				metricType:  "counter",
				metricName:  "SomeMetric",
				metricValue: "84",
			},
			want: http.StatusOK,
		},
		{
			name: "put_existed_counter",
			fields: fields{
				Metrics: map[string]string{
					"SomeMetric": "1",
				},
			},
			args: args{
				metricType:  "counter",
				metricName:  "SomeMetric",
				metricValue: "4",
			},
			want: http.StatusOK,
		},
		{
			name: "put_wrong_counter",
			fields: fields{
				Metrics: map[string]string{},
			},
			args: args{
				metricType:  "counter",
				metricName:  "SomeMetric",
				metricValue: "84.1",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "wrong_value_in_storage",
			fields: fields{
				Metrics: map[string]string{
					"SomeMetric": "1.3",
				},
			},
			args: args{
				metricType:  "counter",
				metricName:  "SomeMetric",
				metricValue: "8",
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "put_wrong_type",
			fields: fields{
				Metrics: map[string]string{},
			},
			args: args{
				metricType:  "yota",
				metricName:  "SomeMetric",
				metricValue: "84",
			},
			want: http.StatusNotImplemented,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := &MemStorage{
				Metrics: tc.fields.Metrics,
			}
			get := ms.Put(tc.args.metricType, tc.args.metricName, tc.args.metricValue)
			assert.Equal(t, tc.want, get)
		})
	}
}

func TestMemStorage_Get(t *testing.T) {
	type fields struct {
		Metrics map[string]string
	}
	type want struct {
		value string
		code  int
	}
	type args struct {
		metricType string
		metricName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "wrong_type",
			fields: fields{
				Metrics: map[string]string{},
			},
			args: args{
				metricType: "yota",
				metricName: "SomeMetric",
			},
			want: want{
				value: "",
				code:  http.StatusNotImplemented,
			},
		},
		{
			name: "not_existed_gauge_name",
			fields: fields{
				Metrics: map[string]string{},
			},
			args: args{
				metricType: "gauge",
				metricName: "SomeMetric",
			},
			want: want{
				value: "",
				code:  http.StatusNotFound,
			},
		},
		{
			name: "not_existed_counter_name",
			fields: fields{
				Metrics: map[string]string{},
			},
			args: args{
				metricType: "counter",
				metricName: "SomeMetric",
			},
			want: want{
				value: "",
				code:  http.StatusNotFound,
			},
		},
		{
			name: "get_gauge",
			fields: fields{
				Metrics: map[string]string{
					"SomeMetric": "4.1",
				},
			},
			args: args{
				metricType: "gauge",
				metricName: "SomeMetric",
			},
			want: want{
				value: "4.1",
				code:  http.StatusOK,
			},
		},
		{
			name: "get_counter",
			fields: fields{
				Metrics: map[string]string{
					"SomeMetric": "4",
				},
			},
			args: args{
				metricType: "counter",
				metricName: "SomeMetric",
			},
			want: want{
				value: "4",
				code:  http.StatusOK,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := &MemStorage{
				Metrics: tc.fields.Metrics,
			}
			getValue, getCode := ms.Get(tc.args.metricType, tc.args.metricName)
			assert.Equal(t, tc.want.value, getValue)
			assert.Equal(t, tc.want.code, getCode)
		})
	}
}

func TestMemStorage_GetAll(t *testing.T) {
	type fields struct {
		Metrics map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "no_values",
			fields: fields{
				Metrics: map[string]string{},
			},
			want: map[string]string{},
		},
		{
			name: "have_values",
			fields: fields{
				Metrics: map[string]string{
					"SomeMetric":    "4.1",
					"AnotherMetric": "3",
				},
			},
			want: map[string]string{
				"SomeMetric":    "4.1",
				"AnotherMetric": "3",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms := &MemStorage{
				Metrics: tc.fields.Metrics,
			}
			get := ms.GetAll()
			assert.Equal(t, tc.want, get)
		})
	}
}

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want interfaces.MetricStorage
	}{
		{
			name: "storage_created",
			want: &MemStorage{map[string]string{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NewMemStorage(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewMemStorage() = %v, want %v", got, tc.want)
			}
		})
	}
}
