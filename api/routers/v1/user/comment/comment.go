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
	"miaoverse/util/pagination"
)

const (
	maxCommentLen        = 1000
	maxConversationDepth = 10 // 楼中楼对话最大收集层数（一般 2-3 层即可覆盖）
)

// CreateHandler 给动态发送评论。拉黑校验由 RequireNoBlock 中间件完成，评论权限在此校验。
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

	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
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
	created, err := servants.InteractsServant.CreateCommentAndMeta(comment, moment.ID, moment.UserID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.CommentCreated(ctx, toCommentInfo(created, moment.ID))
}

// ReplyHandler 回复动态下的评论（楼中楼）。拉黑/被拉黑校验由 RequireNoBlock 中间件完成，
// 评论权限按所属动态的评论权限校验，并写入互动记录（type=reply）。
func ReplyHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &commentreq.ReplyComment{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}

	content := strings.TrimSpace(req.Content)
	if content == "" || len(content) > maxCommentLen {
		return resp.BadRequest(ctx)
	}

	replied, ok := middleware.BlockComment(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
	}
	root, ok := middleware.BlockCommentRoot(ctx)
	if !ok {
		return resp.ServerError(ctx)
	}
	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.ServerError(ctx)
	}

	if err := checkCommentPermission(ctx, servants, uid, moment); err != nil {
		return err
	}

	comment := modelinteracts.Comment{
		UserID:     uid,
		TargetID:   replied.ID,
		TargetType: modelinteracts.CommentTargetComment,
		Content:    content,
		Status:     modelinteracts.CommentStatusNormal,
	}
	interact := modelinteracts.Interacts{
		UserFrom:    uid,
		UserTo:      replied.UserID,
		TargetID:    replied.ID,
		ReferenceID: root.ID,
		Type:        modelinteracts.InteractTypeReply,
		TargetType:  modelinteracts.InteractTargetComment,
		Status:      modelinteracts.InteractStatusNormal,
	}
	created, err := servants.InteractsServant.CreateReplyCommentAndInteract(comment, interact)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.ReplyCreated(ctx, toReplyInfo(created, moment.ID, replied.UserID))
}

// ConversationHandler 获取楼中楼完整对话：传入楼中楼首条评论 id，返回该评论及全部子孙回复。
func ConversationHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if _, ok := middleware.CurrentUID(ctx); !ok {
		return resp.Unauthorized(ctx)
	}

	root, ok := middleware.BlockComment(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
	}
	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.ServerError(ctx)
	}

	offset, limit, ok := pagination.Parse(ctx.Query("offset"), ctx.Query("limit"))
	if !ok {
		return resp.BadRequest(ctx)
	}

	replies, err := servants.InteractsServant.QueryCommentRepliesByRoot(root.ID, maxConversationDepth)
	if err != nil {
		return resp.ServerError(ctx)
	}

	authorByID := map[uint64]uint32{root.ID: root.UserID}
	for i := range replies {
		authorByID[replies[i].ID] = replies[i].UserID
	}

	start := offset
	if start > len(replies) {
		start = len(replies)
	}
	end := start + limit
	if end > len(replies) {
		end = len(replies)
	}

	items := make([]resp.ReplyInfo, 0, end-start)
	for i := start; i < end; i++ {
		items = append(items, toReplyInfo(&replies[i], moment.ID, authorByID[replies[i].TargetID]))
	}

	return resp.Conversation(ctx, resp.ConversationInfo{
		Root:    toCommentInfo(root, moment.ID),
		Count:   int64(len(replies)),
		Replies: items,
	})
}

func toCommentInfo(c *modelinteracts.Comment, momentID uint64) resp.CommentInfo {
	return resp.CommentInfo{
		ID:        c.ID,
		UserID:    c.UserID,
		MomentID:  momentID,
		Content:   c.Content,
		Status:    c.Status,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toReplyInfo(c *modelinteracts.Comment, momentID uint64, replyToUserID uint32) resp.ReplyInfo {
	return resp.ReplyInfo{
		ID:            c.ID,
		UserID:        c.UserID,
		MomentID:      momentID,
		ReplyToID:     c.TargetID,
		ReplyToUserID: replyToUserID,
		Content:       c.Content,
		Status:        c.Status,
		CreatedAt:     c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
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
