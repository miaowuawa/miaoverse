package comment

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/middleware"
	modelinteracts "miaoverse/model/dao/interacts"
	modelmoment "miaoverse/model/dao/moment"
	"miaoverse/model/dto/comment/commentreq"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
)

const (
	maxCommentLen = 1000
)

// CreateHandler 给动态发送评论。被拉黑/拉黑对方或评论权限不允许时拒绝。
func CreateHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &commentreq.CreateComment{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}

	content := strings.TrimSpace(req.Content)
	if req.MomentID == 0 || content == "" || len(content) > maxCommentLen {
		return resp.BadRequest(ctx)
	}

	moment, err := servants.ContentServant.QueryMomentByID(req.MomentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.FileNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}
	if moment.Status != modelmoment.MomentStatusNormal {
		return resp.FileNotFound(ctx)
	}

	blocked, err := servants.BlockServant.IsBlockedEither(ctx.Context(), uid, moment.UserID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if blocked {
		return resp.Blocked(ctx)
	}

	if err := checkCommentPermission(ctx, servants, uid, moment); err != nil {
		return err
	}

	comment := modelinteracts.Comment{
		UserID:     uid,
		TargetID:   moment.ID,
		TargetType: modelinteracts.CommentTargetMoment,
		Content:    content,
		Status:     modelinteracts.CommentStatusNormal,
	}
	created, err := servants.InteractsServant.CreateCommentAndMeta(comment, moment.ID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.CommentCreated(ctx, resp.CommentInfo{
		ID:        created.ID,
		UserID:    created.UserID,
		MomentID:  created.TargetID,
		Content:   created.Content,
		Status:    created.Status,
		CreatedAt: created.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// checkCommentPermission 按动态的评论权限校验：
// 0 全部可评论；1 仅好友（互相关注）；2 仅粉丝（动态作者关注了评论者）；3 全部不可评论。
func checkCommentPermission(ctx fiber.Ctx, servants *server.Servants, uid uint32, m *modelmoment.Moment) error {
	switch m.CommentPermission {
	case modelmoment.MomentCommentPermissionAll:
		return nil
	case modelmoment.MomentCommentPermissionFriends:
		following, err := servants.InteractsServant.IsFollowing(uid, m.UserID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ServerError(ctx)
		}
		if !following {
			return resp.FileNotShared(ctx)
		}
		followedBy, err := servants.InteractsServant.IsFollowing(m.UserID, uid)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ServerError(ctx)
		}
		if !followedBy {
			return resp.FileNotShared(ctx)
		}
		return nil
	case modelmoment.MomentCommentPermissionFans:
		following, err := servants.InteractsServant.IsFollowing(m.UserID, uid)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ServerError(ctx)
		}
		if !following {
			return resp.FileNotShared(ctx)
		}
		return nil
	case modelmoment.MomentCommentPermissionNone:
		return resp.FileNotShared(ctx)
	default:
		return resp.BadRequest(ctx)
	}
}
