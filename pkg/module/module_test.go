package module

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoModRootPathAndName(t *testing.T) {
	t.Parallel()

	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				dir: func() string {
					dir, err := os.Getwd()
					if err != nil {
						t.Fatal(err)
					}
					return dir
				}(),
			},
			want:    "github.com/incu6us/goimports-reviser/v3",
			wantErr: false,
		},
		{
			name: "path is not set error",
			args: args{
				dir: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "path with '.'",
			args: args{
				dir: ".",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			goModRootPath, err := GoModRootPath(tt.args.dir)
			if err != nil && !tt.wantErr {
				t.Errorf("GoModRootPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := Name(goModRootPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Name() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Name() path = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		prepareFn func() string
		want      string
		wantErr   bool
	}{
		{
			name: "read empty go.mod",
			prepareFn: func() string {
				dir := t.TempDir()
				_, err := os.Create(filepath.Join(dir, "go.mod"))
				if err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "check failed parsing of go.mod",
			prepareFn: func() string {
				dir := t.TempDir()
				file, err := os.Create(filepath.Join(dir, "go.mod"))
				if err != nil {
					t.Fatal(err)
				}

				if _, err := file.WriteString("mod test"); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			goModRootPath := tt.prepareFn()
			got, err := Name(goModRootPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Name() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Name() got = %v, want %v", got, tt.want)
			}
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
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success with auto determining",
			args: args{
				projectName: "",
				filePath: func() string {
					dir, err := os.Getwd()
					if err != nil {
						t.Fatal(err)
					}
					return filepath.Join(dir, "module.go")
				}(),
			},
			want:    "github.com/incu6us/goimports-reviser/v3",
			wantErr: false,
		},

		{
			name: "success with manual set",
			args: args{
				projectName: "github.com/incu6us/goimports-reviser/v3",
				filePath:    "",
			},
			want:    "github.com/incu6us/goimports-reviser/v3",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := DetermineProjectName(tt.args.projectName, tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetermineProjectName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetermineProjectName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
