package UserMoment

import (
	"strings"

	modelmoment "miaoverse/model/dao/moment"
	"miaoverse/model/dto/moment/publishreq"
	"miaoverse/model/dto/resp"
)

const (
	maxMomentContentLen = 5000
	maxMomentTitleLen   = 255
)

// NormalizePublish 校验并归一化发布动态请求，返回可入库的动态记录
func NormalizePublish(req *publishreq.PublishMoment) (*modelmoment.Moment, bool) {
	if req == nil {
		return nil, false
	}

	content := strings.TrimSpace(req.Content)
	title := strings.TrimSpace(req.Title)
	if content == "" || len(content) > maxMomentContentLen {
		return nil, false
	}
	if len(title) > maxMomentTitleLen {
		return nil, false
	}

	status := req.Status
	if status != modelmoment.MomentStatusNormal && status != modelmoment.MomentStatusDraft {
		return nil, false
	}

	permission := req.Permission
	if permission > modelmoment.MomentPermissionFans {
		return nil, false
	}

	commentPermission := req.CommentPermission
	if commentPermission > modelmoment.MomentCommentPermissionNone {
		return nil, false
	}

	// 置顶：仅允许个人置顶（1），全站置顶（100）不允许普通用户设置
	top := req.Top
	if top != modelmoment.MomentTopNone && top != modelmoment.MomentTopPersonal {
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
