package helper

import (
	"os"
	"testing"

	"github.com/incu6us/goimports-reviser/v3/reviser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			name: "success with manual filepath",
			args: args{
				projectName: "",
				filePath: func() string {
					dir, err := os.Getwd()
					require.NoError(t, err)
					return dir
				}(),
			},
			want: "github.com/incu6us/goimports-reviser/v3",
		},
		{
			name: "success with stdin",
			args: args{
				projectName: "",
				filePath:    reviser.StandardInput,
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
