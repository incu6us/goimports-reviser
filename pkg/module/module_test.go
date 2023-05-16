package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoModRootPathAndName(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		dir, err := os.Getwd()
		require.NoError(t, err)

		goModRootPath, err := GoModRootPath(dir)
		require.NoError(t, err)

		got, err := Name(goModRootPath)
		require.NoError(t, err)
		assert.Equal(t, "github.com/incu6us/goimports-reviser/v3", got)
	})

	t.Run("path is not set error", func(t *testing.T) {
		t.Parallel()

		goModPath, err := GoModRootPath("")
		assert.Error(t, err)
		assert.Empty(t, goModPath)
	})

	t.Run("path is empty", func(t *testing.T) {
		t.Parallel()

		goModPath, err := GoModRootPath(".")
		assert.NoError(t, err)

		got, err := Name(goModPath)
		assert.Error(t, err)
		assert.Empty(t, got)
	})
}

func TestName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		prepareFn func() string
	}{
		{
			name: "read empty go.mod",
			prepareFn: func() string {
				dir := t.TempDir()
				f, err := os.Create(filepath.Join(dir, "go.mod"))
				require.NoError(t, err)
				require.NoError(t, f.Close())
				return dir
			},
		},
		{
			name: "check failed parsing of go.mod",
			prepareFn: func() string {
				dir := t.TempDir()
				file, err := os.Create(filepath.Join(dir, "go.mod"))
				require.NoError(t, err)
				_, err = file.WriteString("mod test")
				require.NoError(t, err)
				require.NoError(t, file.Close())
				return dir
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			goModRootPath := tt.prepareFn()
			got, err := Name(goModRootPath)
			require.Error(t, err)
			assert.Empty(t, got)
		})
	}
}

func TestDetermineProjectName(t *testing.T) {
	t.Parallel()

	type args struct {
		projectName string
		filePath    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success with auto determining",
			args: args{
				projectName: "",
				filePath: func() string {
					dir, err := os.Getwd()
					require.NoError(t, err)
					return filepath.Join(dir, "module.go")
				}(),
			},
			want: "github.com/incu6us/goimports-reviser/v3",
		},

		{
			name: "success with manual set",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser/v3",
				filePath:    "",
			},
			want: "github.com/incu6us/goimports-reviser/v3",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := DetermineProjectName(tt.args.projectName, tt.args.filePath)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
