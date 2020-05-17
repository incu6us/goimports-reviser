package reviser

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	type args struct {
		projectName string
		filePath    string
		fileContent string
	}
	tests := []struct {
		name       string
		args       args
		want       string
		wantChange bool
		wantErr    bool
	}{
		{
			name: "success with comments",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"log"

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"

	"bytes"

	"github.com/pkg/errors"
)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"bytes"
	"log"

	"github.com/pkg/errors"

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "success with std & project deps",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"log"

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"

	"bytes"


)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"bytes"
	"log"

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "success with std & third-party deps",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata
		
import (
"log"

"bytes"

"github.com/pkg/errors"
)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"bytes"
	"log"

	"github.com/pkg/errors"
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "success with std deps only",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata
		
import (
"log"

"bytes"
)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"bytes"
	"log"
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "success with third-party deps only",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (

	"github.com/pkg/errors"
)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"github.com/pkg/errors"
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "success with project deps only",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"

)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "success with clear doc for import",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"fmt"


	// test
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"fmt"

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "success with comment for import",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"github.com/incu6us/goimports-reviser/testdata/innderpkg" // test1
	
	"fmt" //test2
	// this should be skipped
)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"fmt" // test2

	"github.com/incu6us/goimports-reviser/testdata/innderpkg" // test1
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "success with no changes",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)

// nolint:gomnd
`,
			},
			want: `package testdata

import (
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)

// nolint:gomnd
`,
			wantChange: false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		if err := ioutil.WriteFile(tt.args.filePath, []byte(tt.args.fileContent), 0644); err != nil {
			t.Errorf("write test file failed: %s", err)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, hasChange, err := Execute(tt.args.projectName, tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestExecute_WithRemoveUnusedImports(t *testing.T) {
	type args struct {
		projectName string
		filePath    string
		fileContent string
	}
	tests := []struct {
		name       string
		args       args
		want       string
		wantChange bool
		wantErr    bool
	}{
		{
			name: "remove unused import",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"fmt" //fmt package
	"github.com/incu6us/goimports-reviser/testdata/innderpkg" //custom package
)

// nolint:gomnd
func main(){
  _ = fmt.Println("test")
}
`,
			},
			want: `package testdata

import (
	"fmt" // fmt package
)

// nolint:gomnd
func main() {
	_ = fmt.Println("test")
}
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "remove unused import with alias",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"fmt" //fmt package
	p "github.com/incu6us/goimports-reviser/testdata/innderpkg" //p package
)

// nolint:gomnd
func main(){
  _ = fmt.Println("test")
}
`,
			},
			want: `package testdata

import (
	"fmt" // fmt package
)

// nolint:gomnd
func main() {
	_ = fmt.Println("test")
}
`,
			wantChange: true,
			wantErr:    false,
		},

		{
			name: "use loaded import but not used",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"fmt" //fmt package
	_ "github.com/incu6us/goimports-reviser/testdata/innderpkg" //custom package
)

// nolint:gomnd
func main(){
  _ = fmt.Println("test")
}
`,
			},
			want: `package testdata

import (
	"fmt" // fmt package

	_ "github.com/incu6us/goimports-reviser/testdata/innderpkg" // custom package
)

// nolint:gomnd
func main() {
	_ = fmt.Println("test")
}
`,
			wantChange: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		if err := ioutil.WriteFile(tt.args.filePath, []byte(tt.args.fileContent), 0644); err != nil {
			t.Errorf("write test file failed: %s", err)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, hasChange, err := Execute(tt.args.projectName, tt.args.filePath, OptionRemoveUnusedImports)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestExecute_WithAliasForVersionSuffix(t *testing.T) {
	type args struct {
		projectName string
		filePath    string
		fileContent string
	}
	tests := []struct {
		name       string
		args       args
		want       string
		wantChange bool
		wantErr    bool
	}{
		{
			name: "success with set alias",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package main
import(
	"fmt"
	"github.com/go-pg/pg/v9"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}`,
			},
			want: `package main

import (
	"fmt"
	"strconv"

	pg "github.com/go-pg/pg/v9"
)

func main() {
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}
`,
			wantChange: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		if err := ioutil.WriteFile(tt.args.filePath, []byte(tt.args.fileContent), 0644); err != nil {
			t.Errorf("write test file failed: %s", err)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, hasChange, err := Execute(tt.args.projectName, tt.args.filePath, OptionUseAliasForVersionSuffix)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
