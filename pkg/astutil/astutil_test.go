package astutil

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsesImport(t *testing.T) {
	type args struct {
		fileData string
		path     string
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
			},
			want: true,
		},
	}
	for _, tt := range tests {

		fileData := tt.args.fileData

		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", []byte(fileData), parser.ParseComments)
			if err != nil {
				require.Nil(t, err)
			}

			if got := UsesImport(f, tt.args.path); got != tt.want {
				t.Errorf("UsesImport() = %v, want %v", got, tt.want)
			}
		})
	}
}
