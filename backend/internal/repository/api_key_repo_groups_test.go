package repository

import (
	"reflect"
	"testing"
)

func TestRewriteAPIKeyGroupIDs(t *testing.T) {
	tests := []struct {
		name       string
		groupIDs   []int64
		oldGroupID int64
		newGroupID int64
		want       []int64
	}{
		{name: "remove and compact", groupIDs: []int64{1, 2, 3}, oldGroupID: 2, want: []int64{1, 3}},
		{name: "replace in place", groupIDs: []int64{1, 2, 3}, oldGroupID: 2, newGroupID: 4, want: []int64{1, 4, 3}},
		{name: "replace and deduplicate", groupIDs: []int64{1, 2, 3}, oldGroupID: 2, newGroupID: 3, want: []int64{1, 3}},
		{name: "leave unrelated chain", groupIDs: []int64{1, 2}, oldGroupID: 3, newGroupID: 4, want: []int64{1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rewriteAPIKeyGroupIDs(tt.groupIDs, tt.oldGroupID, tt.newGroupID)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("rewriteAPIKeyGroupIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
