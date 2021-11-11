package grouporder

import (
	"reflect"
	"testing"
)

func TestNewImportGroupOrder(t *testing.T) {
	tests := map[string]struct {
		str    string
		expErr bool
	}{
		"valid": {
			str:    "std,ext,org,prj",
			expErr: false,
		},
		"not enough args": {
			str:    "std,ext,org",
			expErr: true,
		},
		"wrong arg": {
			str:    "std,ext,org,unk",
			expErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewImportGroupOrder(tt.str)
			if (err != nil) != tt.expErr {
				t.Errorf("NewImportGroupOrder() error = %v, wantErr %v", err, tt.expErr)
				return
			}
		})
	}
}

func TestImportGroupOrderGroupedImports(t *testing.T) {
	tests := map[string]struct {
		o        ImportGroupOrder
		std      []string
		org      []string
		project  []string
		external []string
		exp      [][]string
	}{
		"default": {
			o:        []Group{GroupStd, GroupExternal, GroupOrganization, GroupProject},
			std:      []string{"1", "2", "3"},
			org:      []string{"4"},
			project:  []string{"5"},
			external: []string{"6", "7"},
			exp: [][]string{
				{"1", "2", "3"},
				{"6", "7"},
				{"4"},
				{"5"},
			},
		},
		"reverse": {
			o:        []Group{GroupProject, GroupOrganization, GroupExternal, GroupStd},
			std:      []string{"1", "2", "3"},
			org:      []string{"4"},
			project:  []string{"5"},
			external: []string{"6", "7"},
			exp: [][]string{
				{"5"},
				{"4"},
				{"6", "7"},
				{"1", "2", "3"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.o.GroupedImports(tt.std, tt.org, tt.project, tt.external); !reflect.DeepEqual(got, tt.exp) {
				t.Errorf("GroupedImports() = %v, want %v", got, tt.exp)
			}
		})
	}
}
