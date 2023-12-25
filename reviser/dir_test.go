package reviser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sep = string(os.PathSeparator)

func TestNewSourceDir(t *testing.T) {
	t.Run("should generate source dir from recursive path", func(tt *testing.T) {
		dir := NewSourceDir("project", recursivePath, false, "")
		assert.Equal(tt, "project", dir.projectName)
		assert.NotContains(tt, dir.dir, "/...")
		assert.Equal(tt, true, dir.isRecursive)
		assert.Equal(tt, 0, len(dir.excludePatterns))
	})
}

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
		sortedData, _, changed, err := NewSourceFile("testdata", testFile).Fix()
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

func TestSourceDir_Find(t *testing.T) {
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
		sortedData, _, changed, err := NewSourceFile("testdata", testFile).Fix()
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
		want []string
	}{
		{
			name: "found-unformatted",
			args: args{project: "testdata", path: "testdata/dir", excludes: "dir1" + sep + "test.go"},
			want: []string{"testdata/dir/dir1/file1.go"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			exec(tt, func(ttt *testing.T) error {
				files, err := NewSourceDir(test.args.project, test.args.path, true, test.args.excludes).Find()
				assert.NoError(tt, err)
				rootPath, err := os.Getwd()
				assert.NoError(tt, err)
				var want []string
				for _, w := range test.want {
					want = append(want, filepath.Join(rootPath, w))
				}
				assert.Equal(tt, want, files.List())
				return nil
			})
		})
	}
}

func TestUnformattedCollection_List(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *UnformattedCollection
		inspect func(r *UnformattedCollection, t *testing.T) //inspects receiver after test run

		want1 []string
	}{
		{
			name: "sucess",
			init: func(t *testing.T) *UnformattedCollection {
				return newUnformattedCollection([]string{"1", "2"})
			},
			inspect: func(r *UnformattedCollection, t *testing.T) {

			},
			want1: []string{"1", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			got1 := receiver.List()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("UnformattedCollection.List got1 = %v, want1: %v", got1, tt.want1)
			}
		})
	}
}

func TestUnformattedCollection_String(t *testing.T) {
	tests := []struct {
		name    string
		init    func(t *testing.T) *UnformattedCollection
		inspect func(r *UnformattedCollection, t *testing.T) //inspects receiver after test run
		want    string
	}{
		{
			name: "success",
			init: func(t *testing.T) *UnformattedCollection {
				return newUnformattedCollection([]string{"1", "2"})
			},
			inspect: func(r *UnformattedCollection, t *testing.T) {

			},
			want: `1
2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}
			assert.Equal(t, tt.want, receiver.String())
		})
	}
}
