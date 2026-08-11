package Moment

import (
	"errors"

	"gorm.io/gorm"
	"miaoverse/consts"
	"miaoverse/dao/interacts"
	modelmoment "miaoverse/model/dao/moment"
)

// RelationFlags 计算查看者与目标用户的关系：isFriend 互相关注，isFan 目标关注了查看者。
// 与内容列表的可见性判定共用同一套关系计算，避免各接口口径不一致。
func RelationFlags(interactsServant *interacts.InteractsDAO, viewerID uint32, targetID uint32) (bool, bool, error) {
	if viewerID == targetID {
		return false, false, nil
	}

	following, err := interactsServant.IsFollowing(viewerID, targetID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, false, err
	}
	followedBy, err := interactsServant.IsFollowing(targetID, viewerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, false, err
	}
	return following && followedBy, followedBy, nil
}

// VisibleTo 判定动态对查看者是否可见，与 DAO 的 visibleMomentScope 口径一致：
// 公开全部可见；仅好友需互相关注；仅粉丝需作者关注查看者；仅自己仅本人可见。
func VisibleTo(m *modelmoment.Moment, viewerID uint32, isFriend bool, isFan bool) bool {
	if m == nil || m.Status != consts.MomentStatusNormal {
		return false
	}
	if m.Permission == consts.MomentPermissionPublic {
		return true
	}
	if viewerID == m.UserID && m.Permission == consts.MomentPermissionPrivate {
		return true
	}
	if isFriend && m.Permission == consts.MomentPermissionFriends {
		return true
	}
	if isFan && m.Permission == consts.MomentPermissionFans {
		return true
	}
	return false
}
