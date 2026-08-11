package UserMoment

import (
	"testing"

	"miaoverse/consts"
	modelmoment "miaoverse/model/dao/moment"
)

func momentWith(permission uint8, status uint8, userID uint32) *modelmoment.Moment {
	return &modelmoment.Moment{
		ID:         1,
		UserID:     userID,
		Status:     status,
		Permission: permission,
	}
}

func TestVisibleTo(t *testing.T) {
	const (
		author = uint32(10001)
		viewer = uint32(20002)
		other  = uint32(30003)
	)

	cases := []struct {
		name     string
		moment   *modelmoment.Moment
		viewerID uint32
		isFriend bool
		isFan    bool
		want     bool
	}{
		{name: "public visible to anyone", moment: momentWith(consts.MomentPermissionPublic, consts.MomentStatusNormal, author), viewerID: other, want: true},
		{name: "public visible to author", moment: momentWith(consts.MomentPermissionPublic, consts.MomentStatusNormal, author), viewerID: author, want: true},
		{name: "private visible to author only", moment: momentWith(consts.MomentPermissionPrivate, consts.MomentStatusNormal, author), viewerID: author, want: true},
		{name: "private hidden from others", moment: momentWith(consts.MomentPermissionPrivate, consts.MomentStatusNormal, author), viewerID: other, want: false},
		{name: "friends visible to mutual follow", moment: momentWith(consts.MomentPermissionFriends, consts.MomentStatusNormal, author), viewerID: viewer, isFriend: true, want: true},
		{name: "friends hidden without mutual follow", moment: momentWith(consts.MomentPermissionFriends, consts.MomentStatusNormal, author), viewerID: viewer, want: false},
		{name: "fans visible when author follows viewer", moment: momentWith(consts.MomentPermissionFans, consts.MomentStatusNormal, author), viewerID: viewer, isFan: true, want: true},
		{name: "fans hidden when author does not follow viewer", moment: momentWith(consts.MomentPermissionFans, consts.MomentStatusNormal, author), viewerID: viewer, want: false},
		{name: "deleted moment never visible", moment: momentWith(consts.MomentPermissionPublic, consts.MomentStatusDeleted, author), viewerID: other, want: false},
		{name: "draft moment never visible", moment: momentWith(consts.MomentPermissionPublic, consts.MomentStatusDraft, author), viewerID: other, want: false},
		{name: "nil moment never visible", moment: nil, viewerID: other, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := VisibleTo(tc.moment, tc.viewerID, tc.isFriend, tc.isFan); got != tc.want {
				t.Fatalf("VisibleTo = %v, want %v", got, tc.want)
			}
		})
	}
}
