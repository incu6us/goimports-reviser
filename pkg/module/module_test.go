package module

import (
	"os"
	"testing"
)

func TestName(t *testing.T) {
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
