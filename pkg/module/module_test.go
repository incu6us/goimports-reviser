package module

import (
	"os"
	"testing"
)

func TestGoModRootPathAndName(t *testing.T) {
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
						panic(err)
					}

					return dir
				}(),
			},
			want:    "github.com/incu6us/goimports-reviser/v2",
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
		t.Run(tt.name, func(t *testing.T) {
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
	type args struct {
		goModRootPath string
	}
	tests := []struct {
		name      string
		prepareFn func()
		args      args
		want      string
		wantErr   bool
	}{
		{
			name: "read empty go.mod",
			prepareFn: func() {
				const f = "/tmp/go.mod"

				if _, err := os.Stat(f); os.IsExist(err) {
					if err := os.Remove(f); err != nil {
						panic(err)
					}
				}

				_, err := os.Create(f)
				if err != nil {
					panic(err)
				}
			},
			args: args{
				goModRootPath: "/tmp",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "check failed parsing of go.mod",
			prepareFn: func() {
				const f = "/tmp/go.mod"

				if _, err := os.Stat(f); os.IsExist(err) {
					if err := os.Remove(f); err != nil {
						panic(err)
					}
				}

				file, err := os.Create(f)
				if err != nil {
					panic(err)
				}

				if _, err := file.WriteString("mod test"); err != nil {
					panic(err)
				}
			},
			args: args{
				goModRootPath: "/tmp",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.prepareFn()

			got, err := Name(tt.args.goModRootPath)
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
