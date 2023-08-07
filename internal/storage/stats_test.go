package storage

import (
	"fmt"
	"runtime"
	"testing"
)

func TestStatStorage_Update(t *testing.T) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	type fields struct {
		stats map[string]stat
	}
	type args struct {
		memStats runtime.MemStats
		count    int
		rand     float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update_stat",
			fields: fields{
				stats: map[string]stat{
					"Alloc": {
						stype: "gauge",
						name:  "Alloc",
						value: fmt.Sprintf("%v", 844082),
					},
				},
			},
			args: args{
				memStats: ms,
				count:    5,
				rand:     83.2,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tt.fields.stats,
			}
			if err := st.Update(tt.args.memStats, tt.args.count, tt.args.rand); (err != nil) != tt.wantErr {
				t.Errorf("StatStorage.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatStorage_Send(t *testing.T) {
	type fields struct {
		stats map[string]stat
	}
	type args struct {
		url    string
		method string
		action string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "send_statusOK",
			fields: fields{
				stats: map[string]stat{
					"Alloc": {
						stype: "gauge",
						name:  "Alloc",
						value: fmt.Sprintf("%v", 844082),
					},
				},
			},
			args: args{
				url:    "http://localhost:8080",
				method: "POST",
				action: "update",
			},
			want: 200,
		},
		{
			name: "send_wrong_url",
			fields: fields{
				stats: map[string]stat{
					"Alloc": {
						stype: "gauge",
						name:  "Alloc",
						value: fmt.Sprintf("%v", 844082),
					},
				},
			},
			args: args{
				url:    "http://localhost:8089",
				method: "POST",
				action: "update",
			},
			want: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &StatStorage{
				stats: tt.fields.stats,
			}
			if got := st.Send(tt.args.url); got != tt.want {
				t.Errorf("StatStorage.Send() = %v, want %v", got, tt.want)
			}
		})
	}
}
