package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pavlegich/metrics-alerting/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	log := logrus.New()
	h := NewWebhook(log, ms)
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
	log := logrus.New()
	h := NewWebhook(log, ms)
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
	log := logrus.New()
	h := NewWebhook(log, ms)
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
				code:        200,
				contentType: "text/plain",
				body:        "someMetric: 144.1",
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
				code:        404,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:   "/value",
			method: http.MethodGet,
			target: "/value",
			existedValues: map[string]string{
				"someMetric": "144.1",
			},
			want: want{
				code:        400,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:   "without_metric_name",
			method: http.MethodGet,
			target: "/value/gauge",
			existedValues: map[string]string{
				"someMetric": "144.1",
			},
			want: want{
				code:        400,
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
				code:        501,
				contentType: "text/plain",
				body:        "",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h.memStorage = storage.MemStorage{
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
	log := logrus.New()
	h := NewWebhook(log, ms)
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
				code:        200,
				contentType: "text/html; charset=utf-8",
				body: `<html>
	<head>
		<title>Список известных метрик</title>
	</head>
	<body>
		<table>
			<tr>
				<th>Название</th>
				<th>Значение</th>
			</tr>
			<tr>
				<td>someMetric</td>
				<td>144.1</td>
			</tr>
		</table>
	</body>
</html>`,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h.memStorage = storage.MemStorage{
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
