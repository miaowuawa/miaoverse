package comment

import (
	"testing"

	modelinteracts "miaoverse/model/dao/interacts"
)

// TestMaxReplyDepth 覆盖链式回复层数计算：直接回复为 1 层，回复某条回复层数递增，父节点缺失按 1 层处理。
func TestMaxReplyDepth(t *testing.T) {
	const root uint64 = 1

	cases := []struct {
		name    string
		replies []modelinteracts.Comment
		want    int
	}{
		{
			name:    "无回复",
			replies: nil,
			want:    0,
		},
		{
			name: "仅直接回复",
			replies: []modelinteracts.Comment{
				{ID: 2, TargetID: root},
				{ID: 3, TargetID: root},
			},
			want: 1,
		},
		{
			name: "三层链",
			replies: []modelinteracts.Comment{
				{ID: 2, TargetID: root},
				{ID: 3, TargetID: 2},
				{ID: 4, TargetID: 3},
			},
			want: 3,
		},
		{
			name: "取最长分支",
			replies: []modelinteracts.Comment{
				{ID: 2, TargetID: root},
				{ID: 3, TargetID: 2},
				{ID: 4, TargetID: root},
				{ID: 5, TargetID: 3},
				{ID: 6, TargetID: 5},
			},
			want: 4,
		},
		{
			name: "父节点缺失按一层处理",
			replies: []modelinteracts.Comment{
				{ID: 7, TargetID: 999},
				{ID: 8, TargetID: 7},
			},
			want: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := maxReplyDepth(root, tc.replies); got != tc.want {
				t.Fatalf("maxReplyDepth() = %d, want %d", got, tc.want)
			}
		})
	}
}
