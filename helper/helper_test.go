package helper

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/incu6us/goimports-reviser/v3/reviser"
)

func TestDetermineProjectName(t *testing.T) {
	t.Parallel()

	type args struct {
		projectName string
		filePath    string
		option      Option
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
					require.NoError(t, err)
					return dir
				}(),
				option: OSGetwdOption,
			},
			want: "github.com/incu6us/goimports-reviser/v3",
		},
		{
			name: "success with stdin",
			args: args{
				projectName: "",
				filePath:    reviser.StandardInput,
				option:      OSGetwdOption,
			},
			want: "github.com/incu6us/goimports-reviser/v3",
		},
		{
			name: "fail with manual filepath",
			args: args{
				projectName: "",
				filePath:    reviser.StandardInput,
				option: func() (string, error) {
					return "", errors.New("some error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := DetermineProjectName(tt.args.projectName, tt.args.filePath, tt.args.option)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
