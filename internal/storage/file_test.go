package storage

import (
	"context"
	"reflect"
	"testing"
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
		metrics map[string]string
		path    string
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
				metrics: map[string]string{
					"Gauger":  "24.1",
					"Counter": "4",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := NewFile(tt.args.path)
			ms := NewMemStorage(ctx)
			ms.Metrics = tt.args.metrics

			if err := file.Save(ctx, ms); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFile_Load(t *testing.T) {
	ctx := context.Background()
	filePath := "/tmp/metrics-db.json"

	type args struct {
		metrics map[string]string
		path    string
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
				metrics: map[string]string{
					"Gauger":  "24.1",
					"Counter": "4",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := NewFile(tt.args.path)
			ms := NewMemStorage(ctx)
			ms.Metrics = tt.args.metrics

			if err := file.Load(ctx, ms); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
