package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_Set(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{
			name:    "correct_address",
			address: "localhost:8080",
			want:    false,
		},
		{
			name:    "only_host",
			address: "localhost",
			want:    true,
		},
		{
			name:    "only_port",
			address: ":8080",
			want:    true,
		},
		{
			name:    "wrong_port",
			address: "localhost:808o",
			want:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &Address{}
			err := a.Set(tc.address)
			if !tc.want {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}
}

func TestAddress_String(t *testing.T) {
	type fields struct {
		host string
		port int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "localhost:8080",
			fields: fields{
				host: "localhost",
				port: 8080,
			},
			want: "localhost:8080",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := &Address{
				Host: tc.fields.host,
				Port: tc.fields.port,
			}
			assert.Equal(t, tc.want, a.String())
		})
	}
}

func TestAddress_New(t *testing.T) {
	want := &Address{Host: "localhost", Port: 8080}
	assert.Equal(t, want, NewAddress())
}
