package service

import (
	"errors"
	"reflect"
	"testing"
)

func TestNormalizeAPIKeyGroupIDs(t *testing.T) {
	legacy := int64(9)
	tests := []struct {
		name    string
		ids     []int64
		legacy  *int64
		want    []int64
		wantErr error
	}{
		{name: "ordered groups", ids: []int64{3, 1, 2}, want: []int64{3, 1, 2}},
		{name: "legacy group", legacy: &legacy, want: []int64{9}},
		{name: "required", wantErr: ErrAPIKeyGroupRequired},
		{name: "duplicate", ids: []int64{1, 1}, wantErr: ErrAPIKeyDuplicateGroup},
		{name: "invalid", ids: []int64{0}, wantErr: ErrAPIKeyInvalidGroup},
		{name: "too many", ids: []int64{1, 2, 3, 4, 5, 6}, wantErr: ErrAPIKeyTooManyGroups},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAPIKeyGroupIDs(tt.ids, tt.legacy)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("normalizeAPIKeyGroupIDs() error = %v, want %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("normalizeAPIKeyGroupIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
