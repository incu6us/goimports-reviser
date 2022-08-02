package reviser

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceFile_Fix(t *testing.T) {
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
		{
			name: "success no changes by imports and comments",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // configure database/sql with postgres driver
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/incu6us/goimports-reviser/pkg/somepkg"
)
`,
			},
			want: `package testdata

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // configure database/sql with postgres driver
	"github.com/pkg/errors"
	"go.uber.org/fx"

	"github.com/incu6us/goimports-reviser/pkg/somepkg"
)
`,
			wantChange: false,
			wantErr:    false,
		},
		{
			name: "success with multiple import statements",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata

	import "sync" //test comment
	import "testing"

	// yolo
	import "fmt"


	// not sure why this is here but we shall find out soon enough
	import "io"
`,
			},
			want: `package testdata

import (
	"fmt"
	"io"
	"sync" // test comment
	"testing"
)
`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: "preserves cgo import",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/cgo_example.go",
				fileContent: `package testdata

/*
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
)
`,
			},
			want: `package testdata

/*
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
)
`,
			wantChange: false,
			wantErr:    false,
		},
		{
			name: "preserves cgo import even when reordering",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/cgo_example.go",
				fileContent: `package testdata

import (
	"fmt"
	"errors"
)

/*
#include <stdlib.h>
*/
import "C"

import "errors"
`,
			},
			want: `package testdata

import (
	"errors"
	"fmt"
)

/*
#include <stdlib.h>
*/
import "C"
`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: "try to read from stdin",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    StandardInput,
				fileContent: ``,
			},
			wantErr: true,
		},
		{
			name: "error with non-existent file",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdatax/does-not-exist.go",
				fileContent: ``,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if tt.args.filePath != StandardInput && !strings.Contains(tt.args.filePath, "does-not-exist") {
			if err := ioutil.WriteFile(tt.args.filePath, []byte(tt.args.fileContent), 0644); err != nil {
				t.Errorf("write test file failed: %s", err)
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			got, hasChange, err := NewSourceFile(tt.args.projectName, tt.args.filePath).Fix()
			if (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestSourceFile_Fix_WithImportsOrder(t *testing.T) {
	type args struct {
		projectName  string
		filePath     string
		fileContent  string
		importsOrder string
	}

	tests := []struct {
		name       string
		args       args
		want       string
		wantChange bool
		wantErr    bool
	}{
		{
			name: "success with default order",
			args: args{
				projectName:  "github.com/incu6us/goimports-reviser",
				filePath:     "./testdata/example.go",
				importsOrder: "",
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
			name: "success std,general,company,project",
			args: args{
				projectName:  "github.com/incu6us/goimports-reviser",
				filePath:     "./testdata/example.go",
				importsOrder: "std,general,company,project",
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
			name: "success project,company,general,std",
			args: args{
				projectName:  "github.com/incu6us/goimports-reviser",
				filePath:     "./testdata/example.go",
				importsOrder: "project,company,general,std",
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
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"

	"github.com/pkg/errors"

	"bytes"
	"log"
)

// nolint:gomnd
`,
			wantChange: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		if tt.args.filePath != StandardInput && !strings.Contains(tt.args.filePath, "does-not-exist") {
			if err := ioutil.WriteFile(tt.args.filePath, []byte(tt.args.fileContent), 0644); err != nil {
				t.Errorf("write test file failed: %s", err)
			}
		}

		t.Run(tt.name, func(t *testing.T) {
			order, err := StringToImportsOrders(tt.args.importsOrder)
			assert.Nil(t, err)
			got, hasChange, err := NewSourceFile(tt.args.projectName, tt.args.filePath).
				Fix(WithImportsOrder(order))
			if (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestSourceFile_Fix_WithRemoveUnusedImports(t *testing.T) {
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
	"github.com/pkg/errors" //custom package
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
	p "github.com/pkg/errors" //p package
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
	_ "github.com/pkg/errors" //custom package
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

	_ "github.com/pkg/errors" // custom package
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
			name: "success with comments before imports",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `// Some comments are here
package testdata

// test
import (
	"fmt" //fmt package
	_ "github.com/pkg/errors" //custom package
)

// nolint:gomnd
func main(){
  _ = fmt.Println("test")
}
`,
			},
			want: `// Some comments are here
package testdata

// test
import (
	"fmt" // fmt package

	_ "github.com/pkg/errors" // custom package
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
			name: "success without imports",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `// Some comments are here
package testdata

// OutputDir the output directory where the built version of Authelia is located.
var OutputDir = "dist"

// DockerImageName the official name of Authelia docker image.
var DockerImageName = "authelia/authelia"

// IntermediateDockerImageName local name of the docker image.
var IntermediateDockerImageName = "authelia:dist"

const masterTag = "master"
const stringFalse = "false"
const stringTrue = "true"
const suitePathPrefix = "PathPrefix"
const webDirectory = "web"
`,
			},
			want: `// Some comments are here
package testdata

// OutputDir the output directory where the built version of Authelia is located.
var OutputDir = "dist"

// DockerImageName the official name of Authelia docker image.
var DockerImageName = "authelia/authelia"

// IntermediateDockerImageName local name of the docker image.
var IntermediateDockerImageName = "authelia:dist"

const masterTag = "master"
const stringFalse = "false"
const stringTrue = "true"
const suitePathPrefix = "PathPrefix"
const webDirectory = "web"
`,
			wantChange: false,
			wantErr:    false,
		},
		{
			name: "cleanup empty import block",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `// Some comments are here
package testdata

import (
	"fmt"
)

// nolint:gomnd
func main(){
}
`,
			},
			want: `// Some comments are here
package testdata

// nolint:gomnd
func main() {
}
`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: `success with "C"`,
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata
/*
#cgo CFLAGS: -I
#cgo LDFLAGS: -L
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import(
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}`,
			},
			want: `package testdata

/*
#cgo CFLAGS: -I
#cgo LDFLAGS: -L
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"strconv"
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
			got, hasChange, err := NewSourceFile(tt.args.projectName, tt.args.filePath).
				Fix(WithRemovingUnusedImports)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestSourceFile_Fix_WithAliasForVersionSuffix(t *testing.T) {
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
				fileContent: `package testdata
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
			want: `package testdata

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
		{
			name: "success with github.com/pkg/errors",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata
import(
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}`,
			},
			want: `package testdata

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

func main() {
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}
`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: `success with "C"`,
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata
/*
#cgo CFLAGS: -I
#cgo LDFLAGS: -L
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import(
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}`,
			},
			want: `package testdata

/*
#cgo CFLAGS: -I
#cgo LDFLAGS: -L
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
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
			got, hasChange, err := NewSourceFile(tt.args.projectName, tt.args.filePath).
				Fix(WithUsingAliasForVersionSuffix)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestSourceFile_Fix_WithLocalPackagePrefixes(t *testing.T) {
	type args struct {
		projectName      string
		filePath         string
		fileContent      string
		localPkgPrefixes string
	}

	tests := []struct {
		name       string
		args       args
		want       string
		wantChange bool
		wantErr    bool
	}{
		{
			name: "group local packages",
			args: args{
				projectName:      "github.com/incu6us/goimports-reviser",
				localPkgPrefixes: "goimports-reviser",
				filePath:         "./testdata/example.go",
				fileContent: `package testdata

import (
	"fmt" //fmt package
	"github.com/pkg/errors" //custom package
	"github.com/incu6us/goimports-reviser/pkg"
	"goimports-reviser/pkg"
)

/*
#include <stdlib.h>
*/
import "C"

// nolint:gomnd
func main(){
  _ = fmt.Println("test")
}
`,
			},
			want: `package testdata

import (
	"fmt" // fmt package

	"github.com/pkg/errors" // custom package

	"goimports-reviser/pkg"

	"github.com/incu6us/goimports-reviser/pkg"
)

/*
#include <stdlib.h>
*/
import "C"

// nolint:gomnd
func main() {
	_ = fmt.Println("test")
}
`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: "group local packages",
			args: args{
				projectName:      "goimports-reviser",
				localPkgPrefixes: "github.com/incu6us/goimports-reviser",
				filePath:         "./testdata/example.go",
				fileContent: `package testdata

import (
	"fmt" //fmt package
	"github.com/pkg/errors" //custom package
	"github.com/incu6us/goimports-reviser/pkg"
	"goimports-reviser/pkg"
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

	"github.com/pkg/errors" // custom package

	"github.com/incu6us/goimports-reviser/pkg"

	"goimports-reviser/pkg"
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
			name: "group local packages separately from project files",
			args: args{
				projectName:      "github.com/incu6us/goimports-reviser/code/thispkg",
				localPkgPrefixes: "github.com/incu6us/goimports-reviser/code",
				filePath:         "./testdata/example.go",
				fileContent: `package testdata

import (
	"fmt"
	"github.com/3rdparty/pkg"
	"github.com/incu6us/goimports-reviser/code/foopkg"
	"github.com/incu6us/goimports-reviser/code/otherpkg"
	"github.com/incu6us/goimports-reviser/code/thispkg/stuff"
	"github.com/incu6us/goimports-reviser/code/thispkg/morestuff"
)

// nolint:gomnd
func main(){
  _ = fmt.Println("test")
}
`,
			},
			want: `package testdata

import (
	"fmt"

	"github.com/3rdparty/pkg"

	"github.com/incu6us/goimports-reviser/code/foopkg"
	"github.com/incu6us/goimports-reviser/code/otherpkg"

	"github.com/incu6us/goimports-reviser/code/thispkg/morestuff"
	"github.com/incu6us/goimports-reviser/code/thispkg/stuff"
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
			name: "check without local packages",
			args: args{
				projectName:      "github.com/incu6us/goimports-reviser/code/thispkg",
				localPkgPrefixes: "",
				filePath:         "./testdata/example.go",
				fileContent: `package testdata

import (
	"fmt"
	"github.com/3rdparty/pkg"
	"github.com/incu6us/goimports-reviser/code/foopkg"
	"github.com/incu6us/goimports-reviser/code/otherpkg"
	"github.com/incu6us/goimports-reviser/code/thispkg/stuff"
	"github.com/incu6us/goimports-reviser/code/thispkg/morestuff"
)

// nolint:gomnd
func main(){
  _ = fmt.Println("test")
}
`,
			},
			want: `package testdata

import (
	"fmt"

	"github.com/3rdparty/pkg"
	"github.com/incu6us/goimports-reviser/code/foopkg"
	"github.com/incu6us/goimports-reviser/code/otherpkg"

	"github.com/incu6us/goimports-reviser/code/thispkg/morestuff"
	"github.com/incu6us/goimports-reviser/code/thispkg/stuff"
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
			got, hasChange, err := NewSourceFile(tt.args.projectName, tt.args.filePath).
				Fix(WithCompanyPackagePrefixes(tt.args.localPkgPrefixes))
			if (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestSourceFile_Fix_WithFormat(t *testing.T) {
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
			name: "success",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata
type SomeStruct struct{}
type SomeStruct1 struct{}
// SomeStruct2 comments
type SomeStruct2 struct{}
func (s *SomeStruct2) test() {}
func test(){}
func test1(){}
`,
			},
			want: `package testdata

type SomeStruct struct{}

type SomeStruct1 struct{}

// SomeStruct2 comments
type SomeStruct2 struct{}

func (s *SomeStruct2) test() {}

func test() {}

func test1() {}
`,
			wantChange: true,
			wantErr:    false,
		},
		{
			name: "success with comments",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser",
				filePath:    "./testdata/example.go",
				fileContent: `package testdata
// test -  test comment
func test(){}
// test1 -  test comment
func test1(){}
`,
			},
			want: `package testdata

// test -  test comment
func test() {}

// test1 -  test comment
func test1() {}
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
			got, hasChange, err := NewSourceFile(tt.args.projectName, tt.args.filePath).Fix(WithCodeFormatting)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantChange, hasChange)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
