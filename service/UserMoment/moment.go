package UserMoment

import (
	"strings"

	"miaoverse/consts"
	modelmoment "miaoverse/model/dao/moment"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/moment/publishreq"
	"miaoverse/model/dto/resp"
)

// NormalizePublish 校验并归一化发布动态请求，返回可入库的动态记录
func NormalizePublish(req *publishreq.PublishMoment) (*modelmoment.Moment, bool) {
	if req == nil {
		return nil, false
	}

	content := strings.TrimSpace(req.Content)
	title := strings.TrimSpace(req.Title)
	if content == "" || len(content) > consts.MaxMomentContentLen {
		return nil, false
	}
	if len(title) > consts.MaxMomentTitleLen {
		return nil, false
	}

	status := req.Status
	if status != consts.MomentStatusNormal && status != consts.MomentStatusDraft {
		return nil, false
	}

	permission := req.Permission
	if permission > consts.MomentPermissionFans {
		return nil, false
	}

	commentPermission := req.CommentPermission
	if commentPermission > consts.MomentCommentPermissionNone {
		return nil, false
	}

	// 置顶：仅允许个人置顶（1），全站置顶（100）不允许普通用户设置
	top := req.Top
	if top != consts.MomentTopNone && top != consts.MomentTopPersonal {
		return nil, false
	}

	return &modelmoment.Moment{
		Title:             title,
		Content:           content,
		Status:            status,
		Permission:        permission,
		CommentPermission: commentPermission,
		Top:               top,
	}, true
}

func ToMomentInfo(m *modelmoment.Moment) resp.MomentInfo {
	if m == nil {
		return resp.MomentInfo{}
	}
	return resp.MomentInfo{
		ID:                m.ID,
		UserID:            m.UserID,
		Title:             m.Title,
		Content:           m.Content,
		Status:            m.Status,
		Permission:        m.Permission,
		CommentPermission: m.CommentPermission,
		Top:               m.Top,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

// ToContentItem 将动态转换为内容列表项
func ToContentItem(m *modelmoment.Moment, metas map[uint64]modelmoment.MomentInteractCount) resp.ContentItem {
	item := resp.ContentItem{
		ID:   m.ID,
		Type: consts.ContentTypeMoment,
	}
	if meta, ok := metas[m.ID]; ok {
		item.Comment = meta.CommentCount
		item.Like = meta.LikeCount
	}
	return item
}

// ToMomentDetail 将动态组装为详情响应：动态本体 + 作者信息 + 互动计数 + 当前用户互动状态。
// 作者查询失败时不阻断详情返回（作者字段为空，由 handler 决定是否降级），
// 但要求作者信息存在才能返回完整详情，由调用方保证查询前置。
func ToMomentDetail(m *modelmoment.Moment, author *modeluser.User, meta *modelmoment.MomentInteractCount, isLiked bool, isFollowing bool) resp.MomentDetail {
	detail := resp.MomentDetail{
		ID:                m.ID,
		UserID:            m.UserID,
		Title:             m.Title,
		Content:           m.Content,
		Status:            m.Status,
		Permission:        m.Permission,
		CommentPermission: m.CommentPermission,
		Top:               m.Top,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		IsLiked:           isLiked,
		IsFollowing:       isFollowing,
	}
	if author != nil {
		detail.Author = *author
	}
	if meta != nil {
		detail.Stats = resp.MomentStats{
			Likes:    meta.LikeCount,
			Comments: meta.CommentCount,
			Shares:   meta.ShareCount,
		}
	}
	return detail
}
