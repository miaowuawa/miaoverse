package Moment

import (
	"strings"

	"github.com/google/uuid"
	"miaoverse/consts"
	modelmoment "miaoverse/model/dao/moment"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/moment/publishreq"
	"miaoverse/model/dto/moment/updatereq"
	"miaoverse/model/dto/resp"
)

// ValidateImageUUIDs 校验动态图片 UUID 列表：
//   - 数量不超过 MaxMomentImages；
//   - 每个 UUID 必须是合法格式且去重；
//   - 每个文件必须存在、为 active 状态、属于当前用户、且为图片类型（安全：禁止引用他人/非图片/失效文件）。
//
// 返回规范化后的去重 UUID 列表与是否通过校验。
func ValidateImageUUIDs(userServant interface {
	QueryActiveFilesByUUIDsBatch(fileUUIDs []string) (map[string]modeluser.File, error)
}, uid uint32, raw []string) ([]string, bool) {
	if len(raw) == 0 {
		return nil, true
	}
	if len(raw) > consts.MaxMomentImages {
		return nil, false
	}

	seen := make(map[string]bool, len(raw))
	uuids := make([]string, 0, len(raw))
	for _, value := range raw {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false
		}
		if _, err := uuid.Parse(value); err != nil {
			return nil, false
		}
		if seen[value] {
			return nil, false
		}
		seen[value] = true
		uuids = append(uuids, value)
	}

	files, err := userServant.QueryActiveFilesByUUIDsBatch(uuids)
	if err != nil {
		return nil, false
	}
	for _, fileUUID := range uuids {
		file, ok := files[fileUUID]
		if !ok || file.UserID != uid || file.FileType != consts.FileTypeImage {
			return nil, false
		}
	}
	return uuids, true
}

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

// NormalizeUpdate 校验并归一化编辑动态请求，返回可入库的更新字段。
// 仅校验请求中出现的字段（指针非 nil），未出现的字段不修改。
func NormalizeUpdate(req *updatereq.UpdateMoment) (map[string]any, bool) {
	if req == nil {
		return nil, false
	}

	updates := map[string]any{}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if len(title) > consts.MaxMomentTitleLen {
			return nil, false
		}
		updates["title"] = title
	}
	if req.Content != nil {
		content := strings.TrimSpace(*req.Content)
		if content == "" || len(content) > consts.MaxMomentContentLen {
			return nil, false
		}
		updates["content"] = content
	}
	if req.Status != nil {
		status := *req.Status
		if status != consts.MomentStatusNormal && status != consts.MomentStatusDraft {
			return nil, false
		}
		updates["status"] = status
	}
	if req.Permission != nil {
		permission := *req.Permission
		if permission > consts.MomentPermissionFans {
			return nil, false
		}
		updates["permission"] = permission
	}
	if req.CommentPermission != nil {
		commentPermission := *req.CommentPermission
		if commentPermission > consts.MomentCommentPermissionNone {
			return nil, false
		}
		updates["comment_permission"] = commentPermission
	}
	// 置顶：仅允许个人置顶（1），全站置顶（100）不允许普通用户设置
	if req.Top != nil {
		top := *req.Top
		if top != consts.MomentTopNone && top != consts.MomentTopPersonal {
			return nil, false
		}
		updates["top"] = top
	}

	// 请求体为空（无任何可修改字段）时拒绝；仅传 file_uuids（图片替换）时放行
	if len(updates) == 0 && req.FileUUIDs == nil {
		return nil, false
	}
	return updates, true
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
func ToMomentDetail(m *modelmoment.Moment, author *modeluser.User, meta *modelmoment.MomentInteractCount, isLiked bool, isFollowing bool, images []string) resp.MomentDetail {
	detail := resp.MomentDetail{
		ID:                m.ID,
		UserID:            m.UserID,
		Title:             m.Title,
		Content:           m.Content,
		Images:            images,
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
