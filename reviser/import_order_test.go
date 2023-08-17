package reviser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToImportsOrder(t *testing.T) {
	t.Parallel()

	type args struct {
		importsOrder string
	}

	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name:    "invalid groups count",
			args:    args{importsOrder: "std,general"},
			wantErr: `use this parameters to sort all groups of your imports: "std,general,company,project"`,
		},
		{
			name:    "unknown group",
			args:    args{importsOrder: "std,general,company,group"},
			wantErr: `unknown order group type: "group"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := StringToImportsOrders(tt.args.importsOrder)

			assert.Nil(t, got)
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}
