package reviser

import (
	"io/ioutil"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	type args struct {
		projectName	string
		filePath	string
		fileContent	string
	}
	tests := []struct {
		name	string
		args	args
		want	string
		wantErr	bool
	}{
		{
			name:	"success with comments",
			args: args{
				projectName:	"github.com/incu6us/goimports-reviser",
				filePath:	"./testdata/example.go",
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
			wantErr:	false,
		},

		{
			name:	"success with std & project deps",
			args: args{
				projectName:	"github.com/incu6us/goimports-reviser",
				filePath:	"./testdata/example.go",
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
			wantErr:	false,
		},

		{
			name:	"success with std & third-party deps",
			args: args{
				projectName:	"github.com/incu6us/goimports-reviser",
				filePath:	"./testdata/example.go",
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
			wantErr:	false,
		},

		{
			name:	"success with std deps only",
			args: args{
				projectName:	"github.com/incu6us/goimports-reviser",
				filePath:	"./testdata/example.go",
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
			wantErr:	false,
		},

		{
			name:	"success with third-party deps only",
			args: args{
				projectName:	"github.com/incu6us/goimports-reviser",
				filePath:	"./testdata/example.go",
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
			wantErr:	false,
		},

		{
			name:	"success with project deps only",
			args: args{
				projectName:	"github.com/incu6us/goimports-reviser",
				filePath:	"./testdata/example.go",
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
			wantErr:	false,
		},

		{
			name:	"success with no changes",
			args: args{
				projectName:	"github.com/incu6us/goimports-reviser",
				filePath:	"./testdata/example.go",
				fileContent: `package testdata

import (
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)

// nolint:gomnd
`,
			},
			want:		"",
			wantErr:	false,
		},
	}
	for _, tt := range tests {
		if err := ioutil.WriteFile(tt.args.filePath, []byte(tt.args.fileContent), 0644); err != nil {
			t.Errorf("write test file failed: %s", err)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := Execute(tt.args.projectName, tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, string(got))
		})
	}
}
