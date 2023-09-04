package handlers

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pavlegich/metrics-alerting/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ps string = "postgresql://localhost:5432/metrics"

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestCounterPost(t *testing.T) {
	// запуск сервера
	ms := storage.NewMemStorage()
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)
	defer db.Close()
	h := NewWebhook(ms, db)
	ts := httptest.NewServer(h.Route())
	defer ts.Close()

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
	ms := storage.NewMemStorage()
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)
	h := NewWebhook(ms, db)
	ts := httptest.NewServer(h.Route())
	defer ts.Close()

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
	ms := storage.NewMemStorage()
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)
	db.Close()
	h := NewWebhook(ms, db)
	ts := httptest.NewServer(h.Route())
	defer ts.Close()

	type want struct {
		code        int
		contentType string
		body        string
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
			h.MemStorage = &storage.MemStorage{
				Metrics: tc.existedValues,
			}
			resp, get := testRequest(t, ts, tc.method, tc.target)
			defer resp.Body.Close()

			assert.Equal(t, tc.want.code, resp.StatusCode)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tc.want.body, get)
		})
	}
}

func TestMainPage(t *testing.T) {
	// запуск сервера
	ms := storage.NewMemStorage()
	db, err := sql.Open("pgx", ps)
	require.NoError(t, err)
	db.Close()
	h := NewWebhook(ms, db)
	ts := httptest.NewServer(h.Route())
	defer ts.Close()

	type want struct {
		code        int
		contentType string
		body        string
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
			h.MemStorage = &storage.MemStorage{
				Metrics: tc.existedValues,
			}
			resp, get := testRequest(t, ts, tc.method, tc.target)
			defer resp.Body.Close()

			assert.Equal(t, tc.want.code, resp.StatusCode)
			assert.Equal(t, tc.want.contentType, resp.Header.Get("Content-Type"))
			assert.Contains(t, get, tc.want.body)
		})
	}
}
