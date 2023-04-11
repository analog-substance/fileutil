package fileutil

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReadFileLineByLine(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Multi line",
			args: args{
				r: strings.NewReader("line 1\nline 2\nline 3"),
			},
			want: []string{
				"line 1",
				"line 2",
				"line 3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChan := ReadFileLineByLine(tt.args.r)

			var got []string
			for g := range gotChan {
				got = append(got, g)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadFileLineByLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadLineByLine(t *testing.T) {
	tempDir := t.TempDir()

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		setup   func()
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Multi line file",
			setup: func() {
				os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("line 1\nline 2\nline 3"), 0644)
			},
			args: args{
				path: filepath.Join(tempDir, "test.txt"),
			},
			want: []string{
				"line 1",
				"line 2",
				"line 3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			gotChan, err := ReadLineByLine(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadLineByLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var got []string
			for g := range gotChan {
				got = append(got, g)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadLineByLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
