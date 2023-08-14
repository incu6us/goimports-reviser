package reviser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSourceDir_IsExcluded(t *testing.T) {
	type args struct {
		project  string
		path     string
		excludes string
		testPath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "normal",
			args: args{project: "project", path: "/project", excludes: "test.go", testPath: "/project/test.go"},
			want: true,
		},
		{
			name: "dir",
			args: args{project: "project", path: "/project", excludes: "test/", testPath: "/project/test"},
			want: true,
		},
		{
			name: "wildcard-1",
			args: args{project: "project", path: "/project", excludes: "tes?.go", testPath: "/project/test.go"},
			want: true,
		},
		{
			name: "wildcard-2",
			args: args{project: "project", path: "/project", excludes: "t*.go", testPath: "/project/test.go"},
			want: true,
		},
		{
			name: "not-excluded",
			args: args{project: "project", path: "/project", excludes: "t*.go", testPath: "/project/abc.go"},
			want: false,
		},
		{
			name: "multi-excludes",
			args: args{project: "project", path: "/project", excludes: "t*.go,abc.go", testPath: "/project/abc.go"},
			want: true,
		},
	}

	for _, test := range tests {
		args := test.args
		t.Run(test.name, func(tt *testing.T) {
			excluded := NewSourceDir(args.project, args.path, true, args.excludes).isExcluded(args.testPath)
			assert.Equal(tt, test.want, excluded)
		})
	}
}
