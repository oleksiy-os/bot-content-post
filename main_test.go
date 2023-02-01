package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr string
	}{
		{
			name:    "ok",
			args:    args{path: "tests/test-config.json"},
			want:    &Config{BotApiKey: "botApiKey test"},
			wantErr: "",
		},
		{
			name:    "file path not found",
			args:    args{path: "path/not-found"},
			want:    nil,
			wantErr: "open path/not-found: no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig(&tt.args.path)
			if tt.wantErr != "" {
				assert.Equal(t, tt.wantErr, err.Error())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// TODO tests with bot API
