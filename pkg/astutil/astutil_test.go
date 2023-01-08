package astutil

import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsesImport(t *testing.T) {
	type args struct {
		fileData       string
		path           string
		packageImports map[string]string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success with github.com/go-pg/pg/v9",
			args: args{
				fileData: `package main
import(
	"fmt"
	"github.com/go-pg/pg/v9"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}
`,
				path: "github.com/go-pg/pg/v9",
				packageImports: map[string]string{
					"github.com/go-pg/pg/v9": "pg",
				},
			},
			want: true,
		},
		{
			name: `success with "pg2 github.com/go-pg/pg/v9"`,
			args: args{
				fileData: `package main
import(
	"fmt"
	pg2 "github.com/go-pg/pg/v9"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(pg2.In([]string{"test"}))
}
`,
				path: "github.com/go-pg/pg/v9",
			},
			want: true,
		},
		{
			name: "success with strconv",
			args: args{
				fileData: `package main
import(
	"fmt"
	"github.com/go-pg/pg/v9"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}
`,
				path: "strconv",
				packageImports: map[string]string{
					"strconv": "strconv",
				},
			},
			want: true,
		},
		{
			name: "success without ast",
			args: args{
				fileData: `package main
import(
	"fmt"
	"github.com/go-pg/pg/v9"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(pg.In([]string{"test"}))
}
`,
				path: "ast",
			},
			want: false,
		},
		{
			name: "success with github.com/incu6us/goimports-reviser/testdata/innderpkg",
			args: args{
				fileData: `package main
import(
	"fmt"
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
	"strconv"
)

func main(){
	_ = strconv.Itoa(1)
	fmt.Println(innderpkg.Something())
}
`,
				path: "github.com/incu6us/goimports-reviser/testdata/innderpkg",
				packageImports: map[string]string{
					"github.com/incu6us/goimports-reviser/testdata/innderpkg": "innderpkg",
				},
			},
			want: true,
		},
		{
			name: "success with unused strconv",
			args: args{
				fileData: `package main
import(
	"fmt"
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
	"strconv"
)

func main(){
	fmt.Println(innderpkg.Something())
}
`,
				path: "strconv",
			},
			want: false,
		},
	}
	for _, tt := range tests {

		fileData := tt.args.fileData

		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", []byte(fileData), parser.ParseComments)
			require.NoError(t, err)

			got := UsesImport(f, tt.args.packageImports, tt.args.path)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadPackageDeps(t *testing.T) {
	type args struct {
		dir      string
		filename string
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				dir:      "./testdata/",
				filename: "testdata.go",
			},
			want: map[string]string{
				"fmt":                   "fmt",
				"github.com/pkg/errors": "errors",
			},
			wantErr: false,
		},
		{
			name: "success with deprecated build tag",
			args: args{
				dir:      "./testdata/",
				filename: "testdata_with_deprecated_build_tag.go",
			},
			want: map[string]string{
				"fmt":                   "fmt",
				"github.com/pkg/errors": "errors",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := parser.ParseFile(
				token.NewFileSet(),
				fmt.Sprintf("%s/%s", tt.args.dir, tt.args.filename),
				nil,
				parser.ParseComments,
			)
			require.NoError(t, err)

			got, err := LoadPackageDependencies(tt.args.dir, ParseBuildTag(f))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}
