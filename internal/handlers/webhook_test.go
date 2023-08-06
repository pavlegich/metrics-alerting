package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounterHandlers(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name   string
		method string
		target string
		want   want
	}{
		{
			name:   "update",
			method: http.MethodPost,
			target: "/update",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:   "without_id",
			method: http.MethodPost,
			target: "/update/counter//42",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:   "invalid_value",
			method: http.MethodPost,
			target: "/update/counter/someMetric/42.1",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
		{
			name:   "status_ok",
			method: http.MethodPost,
			target: "/update/counter/someMetric/10",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.target, nil)
			w := httptest.NewRecorder()
			Webhook(w, request)

			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			assert.Equal(t, tc.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGaugeHandlers(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name   string
		method string
		target string
		want   want
	}{
		{
			name:   "update",
			method: http.MethodPost,
			target: "/update",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:   "without_id",
			method: http.MethodPost,
			target: "/update/gauge//42",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
		{
			name:   "invalid_value",
			method: http.MethodPost,
			target: "/update/gauge/someMetric/42e",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
		{
			name:   "status_ok",
			method: http.MethodPost,
			target: "/update/gauge/someMetric/10.1",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.target, nil)
			w := httptest.NewRecorder()
			Webhook(w, request)

			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			assert.Equal(t, tc.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestWrongRequests(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name   string
		method string
		target string
		want   want
	}{
		{
			name:   "wrong_method",
			method: http.MethodGet,
			target: "/update",
			want: want{
				code:        405,
				contentType: "text/plain",
			},
		},
		{
			name:   "wrong_metric_type",
			method: http.MethodPost,
			target: "/update/someType/someMetric/30",
			want: want{
				code:        400,
				contentType: "text/plain",
			},
		},
		{
			name:   "wrong_metrics_action",
			method: http.MethodPost,
			target: "/upgrade/gauge/someMetric/30",
			want: want{
				code:        404,
				contentType: "text/plain",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.target, nil)
			w := httptest.NewRecorder()
			Webhook(w, request)

			res := w.Result()
			assert.Equal(t, tc.want.code, res.StatusCode)
			assert.Equal(t, tc.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
