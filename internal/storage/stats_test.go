package storage

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
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
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "update_stat",
			fields: fields{
				stats: map[string]stat{},
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
	want := &StatStorage{stats: make(map[string]stat)}
	assert.Equal(t, want, NewStatsStorage())
}
