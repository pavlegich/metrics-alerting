package storage

import (
	"context"
	"reflect"
	"testing"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

func TestFile_New(t *testing.T) {
	tests := []struct {
		name string
		want *FileMetrics
	}{
		{
			name: "new",
			want: &FileMetrics{make(map[string]string)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFileMetrics(context.Background()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFileMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFile_Save(t *testing.T) {
	ctx := context.Background()
	filePath := "/tmp/metrics-db.json"

	type args struct {
		ms   interfaces.MetricStorage
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: filePath,
				ms: &MemStorage{
					Metrics: map[string]string{
						"Gauger":  "24.1",
						"Counter": "4",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := NewFile(tt.args.path)
			if err := file.Save(ctx, tt.args.ms); (err != nil) != tt.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFile_Load(t *testing.T) {
	ctx := context.Background()
	filePath := "/tmp/metrics-db.json"

	type args struct {
		ms   interfaces.MetricStorage
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				path: filePath,
				ms: &MemStorage{
					Metrics: map[string]string{
						"Gauger":  "24.1",
						"Counter": "4",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := NewFile(tt.args.path)
			if err := file.Load(ctx, tt.args.ms); (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
