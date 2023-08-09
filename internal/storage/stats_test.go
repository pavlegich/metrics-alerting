package storage

// func testRequest(t *testing.T, ts *httptest.Server, method,
// 	path string) (*http.Response, string) {
// 	req, err := http.NewRequest(method, ts.URL+path, nil)
// 	require.NoError(t, err)

// 	resp, err := ts.Client().Do(req)
// 	require.NoError(t, err)
// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	require.NoError(t, err)

// 	return resp, string(respBody)
// }

// func TestStatStorage_Update(t *testing.T) {
// 	var ms runtime.MemStats
// 	runtime.ReadMemStats(&ms)
// 	type fields struct {
// 		stats map[string]stat
// 	}
// 	type args struct {
// 		memStats runtime.MemStats
// 		count    int
// 		rand     float64
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "update_stat",
// 			fields: fields{
// 				stats: map[string]stat{
// 					"Alloc": {
// 						stype: "gauge",
// 						name:  "Alloc",
// 						value: fmt.Sprintf("%v", 844082),
// 					},
// 				},
// 			},
// 			args: args{
// 				memStats: ms,
// 				count:    5,
// 				rand:     83.2,
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			st := &StatStorage{
// 				stats: tt.fields.stats,
// 			}
// 			if err := st.Update(tt.args.memStats, tt.args.count, tt.args.rand); (err != nil) != tt.wantErr {
// 				t.Errorf("StatStorage.Update() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestStatStorage_Send(t *testing.T) {
// 	// запуск сервера
// 	ms := NewMemStorage()
// 	log := logrus.New()
// 	h := handlers.NewWebhook(log, ms)
// 	ts := httptest.NewServer(h.Route())
// 	defer ts.Close()

// 	type fields struct {
// 		stats map[string]stat
// 	}
// 	type args struct {
// 		url    string
// 		method string
// 		action string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   int
// 	}{
// 		{
// 			name: "successful_send",
// 			fields: fields{
// 				stats: map[string]stat{
// 					"Alloc": {
// 						stype: "gauge",
// 						name:  "Alloc",
// 						value: fmt.Sprintf("%v", 844082),
// 					},
// 				},
// 			},
// 			args: args{
// 				url:    "http://localhost:8080",
// 				method: http.MethodPost,
// 				action: "update",
// 			},
// 			want: http.StatusOK,
// 		},
// 	}
// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			st := &StatStorage{
// 				stats: tc.fields.stats,
// 			}
// 			testRequest(t, ts, tc.args.method, )
// 			status := st.Send(tc.args.url)
// 			// создаём новый Recorder
// 			w := httptest.NewRecorder()
// 			StatusHandler(w, request)
// 		})
// 	}
// }
