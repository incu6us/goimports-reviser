package helper

import (
	"os"
	"testing"

	"github.com/incu6us/goimports-reviser/v3/reviser"
)

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
			name: "success with manual filepath",
			args: args{
				projectName: "",
				filePath: func() string {
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
			name: "success with stdin",
			args: args{
				projectName: "",
				filePath: func() string {
					return reviser.StandardInput
				}(),
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
