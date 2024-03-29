package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/mocks"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (*http.Response, string) {
	req, err := http.NewRequestWithContext(context.Background(), method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestWebhook_HandleMain(t *testing.T) {
	// запуск сервера
	ctx := context.Background()
	ms := storage.NewMemStorage(ctx)
	cfg := &config.ServerConfig{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mocks.NewMockStorage(ctrl)

	type want struct {
		contentType string
		body        string
		code        int
	}
	tests := []struct {
		name          string
		method        string
		target        string
		existedValues map[string]string
		want          want
	}{
		{
			name:   "main_page",
			method: http.MethodGet,
			target: "/",
			existedValues: map[string]string{
				"someMetric": "144.1",
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/html; charset=utf-8",
				body:        "<td>someMetric</td><td>144.1</td>",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms.Metrics = tc.existedValues

			h := NewWebhook(ctx, ms, mockDB, nil, cfg)
			ts := httptest.NewServer(h.Route(ctx))
			defer ts.Close()

			resp, get := testRequest(t, ts, tc.method, tc.target)
			defer resp.Body.Close()

			assert.Equal(t, tc.want.code, resp.StatusCode)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
			assert.Contains(t, get, tc.want.body)
		})
	}
}

func TestCounterPost(t *testing.T) {
	// запуск сервера
	ctx := context.Background()
	ms := storage.NewMemStorage(ctx)
	cfg := &config.ServerConfig{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mocks.NewMockStorage(ctrl)

	h := NewWebhook(ctx, ms, mockDB, nil, cfg)
	ts := httptest.NewServer(h.Route(ctx))
	defer ts.Close()

	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name   string
		method string
		target string
		want   want
	}{
		{
			name:   "without_id",
			method: http.MethodPost,
			target: "/update/counter//42",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain",
			},
		},
		{
			name:   "invalid_value",
			method: http.MethodPost,
			target: "/update/counter/someMetric/42.1",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
		{
			name:   "status_ok",
			method: http.MethodPost,
			target: "/update/counter/someMetric/10",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, tc.method, tc.target)
			defer resp.Body.Close()

			assert.Equal(t, tc.want.code, resp.StatusCode)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestGaugePost(t *testing.T) {
	// запуск сервера
	ctx := context.Background()
	ms := storage.NewMemStorage(ctx)
	cfg := &config.ServerConfig{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mocks.NewMockStorage(ctrl)

	h := NewWebhook(ctx, ms, mockDB, nil, cfg)
	ts := httptest.NewServer(h.Route(ctx))
	defer ts.Close()

	type want struct {
		contentType string
		code        int
	}
	tests := []struct {
		name   string
		method string
		target string
		want   want
	}{
		{
			name:   "without_id",
			method: http.MethodPost,
			target: "/update/gauge//42",
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain",
			},
		},
		{
			name:   "invalid_value",
			method: http.MethodPost,
			target: "/update/gauge/someMetric/42e",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
		{
			name:   "status_ok",
			method: http.MethodPost,
			target: "/update/gauge/someMetric/10.1",
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, tc.method, tc.target)
			defer resp.Body.Close()

			assert.Equal(t, tc.want.code, resp.StatusCode)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
		})
	}
}

func TestGaugeGet(t *testing.T) {
	// запуск сервера
	ctx := context.Background()
	ms := storage.NewMemStorage(ctx)
	cfg := &config.ServerConfig{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mocks.NewMockStorage(ctrl)

	type want struct {
		contentType string
		body        string
		code        int
	}
	tests := []struct {
		name          string
		method        string
		target        string
		existedValues map[string]string
		want          want
	}{
		{
			name:   "existed_value",
			method: http.MethodGet,
			target: "/value/gauge/someMetric",
			existedValues: map[string]string{
				"someMetric": "144.1",
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
				body:        "144.1",
			},
		},
		{
			name:   "not_existed_value",
			method: http.MethodGet,
			target: "/value/gauge/anotherMetric",
			existedValues: map[string]string{
				"someMetric": "144.1",
			},
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:   "wrong_metric_type",
			method: http.MethodGet,
			target: "/value/yota/someMetric",
			existedValues: map[string]string{
				"someMetric": "144.1",
			},
			want: want{
				code:        http.StatusNotImplemented,
				contentType: "text/plain",
				body:        "",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ms.Metrics = tc.existedValues

			h := NewWebhook(ctx, ms, mockDB, nil, cfg)
			ts := httptest.NewServer(h.Route(ctx))
			defer ts.Close()

			resp, get := testRequest(t, ts, tc.method, tc.target)
			defer resp.Body.Close()

			assert.Equal(t, tc.want.code, resp.StatusCode)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tc.want.body, get)
		})
	}
}
