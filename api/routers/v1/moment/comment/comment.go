package comment

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/consts"
	"miaoverse/middleware"
	modelinteracts "miaoverse/model/dao/interacts"
	modelmoment "miaoverse/model/dao/moment"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/comment/commentreq"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/Sticker"
	"miaoverse/util/pagination"
)

// CreateHandler 给动态发送评论。内容屏蔽校验由 RequireNoContentBlock、拉黑校验由 RequireNoBlockUser 中间件完成，评论权限在此校验。
// 评论内容可内嵌一个贴纸（[sticker:<uuid>] 标记随文本穿插展示），贴纸校验见 checkCommentSticker。
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
	if req.MomentID == 0 || content == "" || len(content) > consts.MaxCommentLen {
		return resp.BadRequest(ctx)
	}

	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
	}

	if err := checkCommentPermission(ctx, servants, uid, moment); err != nil {
		return err
	}

	stickerUUID, err := checkCommentSticker(ctx, servants, uid, content)
	if err != nil {
		return err
	}

	comment := modelinteracts.Comment{
		UserID:      uid,
		TargetID:    moment.ID,
		TargetType:  consts.CommentTargetMoment,
		Content:     content,
		StickerUUID: stickerUUID,
		Status:      consts.CommentStatusNormal,
	}
	created, err := servants.InteractsServant.CreateCommentAndMeta(comment, moment.ID, moment.UserID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	info, err := buildCommentInfo(servants, created, moment.ID, 0, false)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.CommentCreated(ctx, info)
}

// ReplyHandler 回复动态下的评论（楼中楼）。内容屏蔽校验由 RequireNoContentBlock、拉黑/被拉黑校验由 RequireNoBlockUser 中间件完成，
// 评论权限按所属动态的评论权限校验，并写入互动记录（type=reply）。贴纸规则与评论一致。
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
	if content == "" || len(content) > consts.MaxCommentLen {
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

	stickerUUID, err := checkCommentSticker(ctx, servants, uid, content)
	if err != nil {
		return err
	}

	comment := modelinteracts.Comment{
		UserID:      uid,
		TargetID:    replied.ID,
		TargetType:  consts.CommentTargetComment,
		Content:     content,
		StickerUUID: stickerUUID,
		Status:      consts.CommentStatusNormal,
	}
	interact := modelinteracts.Interacts{
		UserFrom:    uid,
		UserTo:      replied.UserID,
		TargetID:    replied.ID,
		ReferenceID: root.ID,
		Type:        consts.InteractTypeReply,
		TargetType:  consts.InteractTargetComment,
		Status:      consts.InteractStatusNormal,
	}
	created, err := servants.InteractsServant.CreateReplyCommentAndInteract(comment, interact)
	if err != nil {
		return resp.ServerError(ctx)
	}

	sticker, err := resolveSticker(servants, created.StickerUUID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.ReplyCreated(ctx, toReplyInfo(created, moment.ID, replied.UserID, sticker))
}

// ListHandler 获取动态一级评论分页列表（需登录）。
// 排序：sort=hot（默认，按点赞数倒序）/ sort=time（按时间倒序）。
// 内容屏蔽校验由 RequireNoContentBlock、拉黑校验由 RequireNoBlockUser 中间件完成（与动态详情一致）。
// 返回的评论 Content 中保留贴纸内嵌标记 [sticker:<uuid>]，Sticker 为贴纸展示信息（Hidden=true 时前端提示「部分贴纸未显示」）。
func ListHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
	}

	offset, limit, ok := pagination.Parse(ctx.Query("offset"), ctx.Query("limit"))
	if !ok {
		return resp.BadRequest(ctx)
	}
	sort := strings.TrimSpace(ctx.Query("sort"))
	if sort == "" {
		sort = consts.CommentSortHot
	}
	if sort != consts.CommentSortHot && sort != consts.CommentSortTime {
		return resp.BadRequest(ctx)
	}

	comments, err := servants.InteractsServant.QueryMomentComments(moment.ID, offset, limit, sort)
	if err != nil {
		return resp.ServerError(ctx)
	}
	count, err := servants.InteractsServant.CountMomentCommentsReal(moment.ID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	commentIDs := make([]uint64, 0, len(comments))
	authorIDs := make([]uint32, 0, len(comments))
	stickerUUIDs := make([]string, 0, len(comments))
	for i := range comments {
		commentIDs = append(commentIDs, comments[i].ID)
		authorIDs = append(authorIDs, comments[i].UserID)
		if comments[i].StickerUUID != "" {
			stickerUUIDs = append(stickerUUIDs, comments[i].StickerUUID)
		}
	}

	authors, err := servants.UserServant.QueryUsersByIDs(authorIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	likes, err := servants.InteractsServant.QueryCommentInteractCountsBatch(commentIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	liked, err := servants.InteractsServant.HasLikedCommentsBatch(uid, commentIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	stickers, err := Sticker.ResolveCommentStickers(servants, stickerUUIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}

	items := make([]resp.CommentInfo, 0, len(comments))
	for i := range comments {
		items = append(items, toCommentInfoWith(&comments[i], moment.ID, authors[comments[i].UserID], likes[comments[i].ID], liked[comments[i].ID], lookupSticker(stickers, comments[i].StickerUUID)))
	}
	return resp.CommentList(ctx, count, items)
}

// LikeCommentHandler 给评论点赞。内容屏蔽校验由 RequireNoContentBlock、拉黑校验由 RequireNoBlockUser 中间件完成。
func LikeCommentHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	comment, ok := middleware.BlockComment(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	if err := servants.InteractsServant.LikeCommentAndMeta(uid, comment.ID, comment.UserID); err != nil {
		return resp.ServerError(ctx)
	}

	return resp.InteractOK(ctx, comment.ID, consts.ActionLike)
}

// UnlikeCommentHandler 取消评论点赞（幂等：未点赞时直接返回成功）。
// 内容屏蔽校验由 RequireNoContentBlock、拉黑校验由 RequireNoBlockUser 中间件完成。
func UnlikeCommentHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	comment, ok := middleware.BlockComment(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	if err := servants.InteractsServant.UnlikeCommentAndMeta(uid, comment.ID); err != nil {
		return resp.ServerError(ctx)
	}

	return resp.InteractOK(ctx, comment.ID, consts.ActionUnlike)
}

// ConversationHandler 获取楼中楼完整对话：传入楼中楼首条评论 id，返回该评论及全部子孙回复。
func ConversationHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
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

	replies, err := servants.InteractsServant.QueryCommentRepliesByRoot(root.ID, consts.MaxConversationDepth)
	if err != nil {
		return resp.ServerError(ctx)
	}

	authorByID := map[uint64]uint32{root.ID: root.UserID}
	for i := range replies {
		authorByID[replies[i].ID] = replies[i].UserID
	}

	// 批量装配点赞状态与贴纸展示信息
	commentIDs := make([]uint64, 0, len(replies)+1)
	commentIDs = append(commentIDs, root.ID)
	stickerUUIDs := make([]string, 0, len(replies)+1)
	if root.StickerUUID != "" {
		stickerUUIDs = append(stickerUUIDs, root.StickerUUID)
	}
	for i := range replies {
		commentIDs = append(commentIDs, replies[i].ID)
		if replies[i].StickerUUID != "" {
			stickerUUIDs = append(stickerUUIDs, replies[i].StickerUUID)
		}
	}
	likes, err := servants.InteractsServant.QueryCommentInteractCountsBatch(commentIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	liked, err := servants.InteractsServant.HasLikedCommentsBatch(uid, commentIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	stickers, err := Sticker.ResolveCommentStickers(servants, stickerUUIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}

	rootInfo, err := buildCommentInfo(servants, root, moment.ID, likes[root.ID], liked[root.ID])
	if err != nil {
		return resp.ServerError(ctx)
	}
	rootInfo.Sticker = lookupSticker(stickers, root.StickerUUID)

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
		items = append(items, toReplyInfo(&replies[i], moment.ID, authorByID[replies[i].TargetID], lookupSticker(stickers, replies[i].StickerUUID)))
	}

	return resp.Conversation(ctx, resp.ConversationInfo{
		Root:    rootInfo,
		Count:   int64(len(replies)),
		Replies: items,
	})
}

// buildCommentInfo 组装评论响应 DTO：附作者、点赞数/点赞状态与贴纸展示信息。
func buildCommentInfo(servants *server.Servants, c *modelinteracts.Comment, momentID uint64, likes uint32, liked bool) (resp.CommentInfo, error) {
	author, err := servants.UserServant.QueryByID(c.UserID)
	if err != nil {
		return resp.CommentInfo{}, err
	}
	sticker, err := resolveSticker(servants, c.StickerUUID)
	if err != nil {
		return resp.CommentInfo{}, err
	}
	return toCommentInfoWith(c, momentID, *author, likes, liked, sticker), nil
}

func toCommentInfoWith(c *modelinteracts.Comment, momentID uint64, author modeluser.User, likes uint32, liked bool, sticker *resp.CommentSticker) resp.CommentInfo {
	return resp.CommentInfo{
		ID:        c.ID,
		UserID:    c.UserID,
		MomentID:  momentID,
		Content:   c.Content,
		Status:    c.Status,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
		Author:    author,
		Likes:     likes,
		IsLiked:   liked,
		Sticker:   sticker,
	}
}

func toReplyInfo(c *modelinteracts.Comment, momentID uint64, replyToUserID uint32, sticker *resp.CommentSticker) resp.ReplyInfo {
	return resp.ReplyInfo{
		ID:            c.ID,
		UserID:        c.UserID,
		MomentID:      momentID,
		ReplyToID:     c.TargetID,
		ReplyToUserID: replyToUserID,
		Content:       c.Content,
		Status:        c.Status,
		CreatedAt:     c.CreatedAt.Format("2006-01-02 15:04:05"),
		Sticker:       sticker,
	}
}

// resolveSticker 解析单条评论的贴纸展示信息（无贴纸返回 nil）。
func resolveSticker(servants *server.Servants, stickerUUID string) (*resp.CommentSticker, error) {
	if stickerUUID == "" {
		return nil, nil
	}
	resolved, err := Sticker.ResolveCommentStickers(servants, []string{stickerUUID})
	if err != nil {
		return nil, err
	}
	return lookupSticker(resolved, stickerUUID), nil
}

// lookupSticker 从批量解析结果取贴纸展示信息（无贴纸返回 nil）。
func lookupSticker(resolved map[string]resp.CommentSticker, stickerUUID string) *resp.CommentSticker {
	if stickerUUID == "" {
		return nil
	}
	info, ok := resolved[strings.ToLower(stickerUUID)]
	if !ok {
		return nil
	}
	return &info
}

// checkCommentSticker 校验评论内容中的贴纸使用并返回贴纸 UUID（无贴纸返回空串）。
// 产品规则：登录且绑定手机号的用户可使用贴纸；一条评论最多一个贴纸；
// 仅可使用本人上传/已加入收藏夹/已收藏贴纸包内且未被封禁的贴纸。
func checkCommentSticker(ctx fiber.Ctx, servants *server.Servants, uid uint32, content string) (string, error) {
	stickerUUID, err := Sticker.CheckCommentSticker(servants, uid, content)
	if err != nil {
		switch {
		case errors.Is(err, Sticker.ErrTokenExceeded):
			return "", resp.StickerInvalid(ctx)
		case errors.Is(err, Sticker.ErrNotFound):
			return "", resp.StickerNotFound(ctx)
		case errors.Is(err, Sticker.ErrPackBanned):
			return "", resp.StickerPackBanned(ctx)
		case errors.Is(err, Sticker.ErrNotUsable):
			return "", resp.StickerNotUsable(ctx)
		default:
			return "", resp.ServerError(ctx)
		}
	}
	if stickerUUID == "" {
		return "", nil
	}

	bound, err := servants.UserServant.HasCredential(uid, consts.Phone)
	if err != nil {
		return "", resp.ServerError(ctx)
	}
	if !bound {
		return "", resp.PhoneNotBound(ctx)
	}
	return stickerUUID, nil
}

// checkCommentPermission 按动态的评论权限校验：
// 0 全部可评论；1 仅好友（互相关注）；2 仅粉丝（动态作者关注了评论者）；3 全部不可评论。
func checkCommentPermission(ctx fiber.Ctx, servants *server.Servants, uid uint32, m *modelmoment.Moment) error {
	switch m.CommentPermission {
	case consts.MomentCommentPermissionAll:
		return nil
	case consts.MomentCommentPermissionFriends:
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
	case consts.MomentCommentPermissionFans:
		following, err := servants.InteractsServant.IsFollowing(m.UserID, uid)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ServerError(ctx)
		}
		if !following {
			return resp.FileNotShared(ctx)
		}
		return nil
	case consts.MomentCommentPermissionNone:
		return resp.FileNotShared(ctx)
	default:
		return resp.BadRequest(ctx)
	}
}
