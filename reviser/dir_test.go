package reviser

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestSourceDir_Fix(t *testing.T) {
	testFile := "testdata/dir/dir1/file1.go"

	originContent := `package dir1
import (
	"strings"
	"fmt"
)
func main() {
	fmt.Println(strings.ToLower("Hello World!"))
}
`
	exec := func(tt *testing.T, fn func(*testing.T) error) {
		// create test file
		err := os.MkdirAll(filepath.Dir(testFile), os.ModePerm)
		assert.NoError(tt, err)
		err = os.WriteFile(testFile, []byte(originContent), os.ModePerm)
		assert.NoError(tt, err)

		// exec test func
		err = fn(tt)
		assert.NoError(tt, err)

		// remove test file
		err = os.Remove(testFile)
		assert.NoError(tt, err)
	}
	var sortedContent string
	exec(t, func(tt *testing.T) error {
		// get sorted content via SourceFile.Fix
		sortedData, changed, err := NewSourceFile("testdata", testFile).Fix()
		assert.NoError(tt, err)
		sortedContent = string(sortedData)
		assert.Equal(tt, true, changed)
		assert.NotEqual(tt, originContent, sortedContent)
		return nil
	})

	type args struct {
		project  string
		path     string
		excludes string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "exclude-file",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1/file1.go"},
			want: originContent,
		}, {
			name: "exclude-dir",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1/"},
			want: originContent,
		}, {
			name: "exclude-file-*",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1/f*1.go"},
			want: originContent,
		}, {
			name: "exclude-file-?",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1/file?.go"},
			want: originContent,
		}, {
			name: "exclude-file-multi",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1/file?.go,aaa,bbb"},
			want: originContent,
		}, {
			name: "not-exclude",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1/test.go"},
			want: sortedContent,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {

			exec(tt, func(ttt *testing.T) error {
				// executing SourceDir.Fix
				err := NewSourceDir(test.args.project, test.args.path, true, test.args.excludes).Fix()
				assert.NoError(tt, err)
				// read new content
				content, err := os.ReadFile(testFile)
				assert.NoError(tt, err)
				assert.Equal(tt, test.want, string(content))
				return nil
			})
		})
	}
}

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
