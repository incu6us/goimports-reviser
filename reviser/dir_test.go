package reviser

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

const sep = string(os.PathSeparator)

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
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1" + sep + "file1.go"},
			want: originContent,
		}, {
			name: "exclude-dir",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1" + sep},
			want: originContent,
		}, {
			name: "exclude-file-*",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1" + sep + "f*1.go"},
			want: originContent,
		}, {
			name: "exclude-file-?",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1" + sep + "file?.go"},
			want: originContent,
		}, {
			name: "exclude-file-multi",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1" + sep + "file?.go,aaa,bbb"},
			want: originContent,
		}, {
			name: "not-exclude",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1" + sep + "test.go"},
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
			args: args{project: "project", path: "project", excludes: "test.go", testPath: "test.go"},
			want: true,
		},
		{
			name: "dir",
			args: args{project: "project", path: "project", excludes: "test/", testPath: "test"},
			want: true,
		},
		{
			name: "wildcard-1",
			args: args{project: "project", path: "project", excludes: "tes?.go", testPath: "test.go"},
			want: true,
		},
		{
			name: "wildcard-2",
			args: args{project: "project", path: "project", excludes: "t*.go", testPath: "test.go"},
			want: true,
		},
		{
			name: "not-excluded",
			args: args{project: "project", path: "project", excludes: "t*.go", testPath: "abc.go"},
			want: false,
		},
		{
			name: "multi-excludes",
			args: args{project: "project", path: "project", excludes: "t*.go,abc.go", testPath: "abc.go"},
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

func join(elem ...string) string {
	return filepath.Join(elem...)
}
