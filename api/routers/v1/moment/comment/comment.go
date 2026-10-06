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
	"miaoverse/service/i18n"
	"miaoverse/util/pagination"
)

// CreateHandler 给动态发送评论。内容屏蔽校验由 RequireNoContentBlock、拉黑校验由 RequireNoBlockUser 中间件完成，评论权限在此校验。
// 评论内容可内嵌贴纸（[sticker:<uuid>] 标记随文本穿插展示，一条评论最多 consts.MaxCommentStickerTokens 张），贴纸校验见 checkCommentStickers。
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
	if req.MomentID == 0 || content == "" || len(Sticker.StripTokens(content)) > consts.MaxCommentLen {
		return resp.BadRequest(ctx)
	}

	moment, ok := middleware.BlockMoment(ctx)
	if !ok {
		return resp.FileNotFound(ctx)
	}

	if err := checkCommentPermission(ctx, servants, uid, moment); err != nil {
		return err
	}

	stickerUUIDs, err := checkCommentStickers(ctx, servants, uid, content)
	if err != nil {
		return err
	}
	stickerUUID := ""
	if len(stickerUUIDs) > 0 {
		stickerUUID = stickerUUIDs[0]
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

	// 被评论通知（旁路业务，失败不影响评论结果；自己评论自己的动态由服务层跳过）
	if actor, ok := middleware.CurrentUser(ctx); ok {
		servants.NotifyServant.NotifyComment(actor, moment.ID, moment.UserID, content, i18n.LanguageFromCtx(ctx))
	}

	info, err := buildCommentInfo(servants, created, moment.ID, 0, false, 0, 0)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.CommentCreated(ctx, info)
}

// ReplyHandler 回复动态下的评论（楼中楼）。内容屏蔽校验由 RequireNoContentBlock、拉黑/被拉黑校验由 RequireNoBlockUser 中间件完成，
// 评论权限按所属动态的评论权限校验，并写入互动记录（type=reply）。贴纸规则与评论一致（可多张，最多 consts.MaxCommentStickerTokens 张）。
// 被回复的评论既可以是楼中楼首条评论，也可以是任意楼中楼回复（回复他人的回复）。
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
	if content == "" || len(Sticker.StripTokens(content)) > consts.MaxCommentLen {
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

	stickerUUIDs, err := checkCommentStickers(ctx, servants, uid, content)
	if err != nil {
		return err
	}
	stickerUUID := ""
	if len(stickerUUIDs) > 0 {
		stickerUUID = stickerUUIDs[0]
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

	// 被回复通知（旁路业务，失败不影响回复结果；自己回复自己由服务层跳过）
	if actor, ok := middleware.CurrentUser(ctx); ok {
		servants.NotifyServant.NotifyReply(actor, replied.ID, replied.UserID, content, i18n.LanguageFromCtx(ctx))
	}

	info, err := buildReplyInfo(servants, created, moment.ID, replied.UserID, 0, false)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.ReplyCreated(ctx, info)
}

// ListHandler 获取动态一级评论分页列表（需登录）。
// 排序：sort=hot（默认，按点赞数倒序）/ sort=time（按时间倒序）。
// 内容屏蔽校验由 RequireNoContentBlock、拉黑校验由 RequireNoBlockUser 中间件完成（与动态详情一致）。
// 返回的评论 Content 中保留贴纸内嵌标记 [sticker:<uuid>]，Stickers 为各标记位置的贴纸展示信息（Hidden=true 时前端提示「部分贴纸未显示」），
// ReplyCount 为该评论楼中楼下的回复总数（含全部子孙回复）。
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
	contents := make([]string, 0, len(comments))
	for i := range comments {
		commentIDs = append(commentIDs, comments[i].ID)
		authorIDs = append(authorIDs, comments[i].UserID)
		contents = append(contents, comments[i].Content)
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
	replyStats, err := servants.InteractsServant.QueryCommentReplyStatsByRootsBatch(commentIDs, consts.MaxConversationDepth)
	if err != nil {
		return resp.ServerError(ctx)
	}
	stickers, err := Sticker.ResolveContentsStickers(servants, contents)
	if err != nil {
		return resp.ServerError(ctx)
	}

	items := make([]resp.CommentInfo, 0, len(comments))
	for i := range comments {
		stats := replyStats[comments[i].ID]
		items = append(items, toCommentInfoWith(&comments[i], moment.ID, authors[comments[i].UserID], likes[comments[i].ID], liked[comments[i].ID], stats.Count, stats.MaxDepth, stickers[i]))
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

	// 被点赞通知（旁路业务，失败不影响点赞结果；自己给自己的评论点赞由服务层跳过）
	if actor, ok := middleware.CurrentUser(ctx); ok {
		servants.NotifyServant.NotifyLikeComment(actor, comment.ID, comment.UserID, comment.Content, i18n.LanguageFromCtx(ctx))
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

	// 被回复评论作者：reply_to_id 为回复链上的上一环（可能是楼中楼首条评论，也可能是任意楼中楼回复）
	replyToUserID := map[uint64]uint32{root.ID: root.UserID}
	for i := range replies {
		replyToUserID[replies[i].ID] = replies[i].UserID
	}

	// 批量装配点赞状态、回复作者与贴纸展示信息
	commentIDs := make([]uint64, 0, len(replies)+1)
	commentIDs = append(commentIDs, root.ID)
	userIDs := make([]uint32, 0, len(replies)+1)
	userIDs = append(userIDs, root.UserID)
	contents := make([]string, 0, len(replies)+1)
	contents = append(contents, root.Content)
	for i := range replies {
		commentIDs = append(commentIDs, replies[i].ID)
		userIDs = append(userIDs, replies[i].UserID)
		contents = append(contents, replies[i].Content)
	}
	likes, err := servants.InteractsServant.QueryCommentInteractCountsBatch(commentIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	liked, err := servants.InteractsServant.HasLikedCommentsBatch(uid, commentIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	authors, err := servants.UserServant.QueryUsersByIDs(userIDs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	stickers, err := Sticker.ResolveContentsStickers(servants, contents)
	if err != nil {
		return resp.ServerError(ctx)
	}

	rootInfo := toCommentInfoWith(root, moment.ID, authors[root.UserID], likes[root.ID], liked[root.ID], int64(len(replies)), maxReplyDepth(root.ID, replies), stickers[0])

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
		items = append(items, toReplyInfoWith(&replies[i], moment.ID, replyToUserID[replies[i].TargetID], authors[replies[i].UserID], likes[replies[i].ID], liked[replies[i].ID], stickers[i+1]))
	}

	return resp.Conversation(ctx, resp.ConversationInfo{
		Root:    rootInfo,
		Count:   int64(len(replies)),
		Replies: items,
	})
}

// buildCommentInfo 组装评论响应 DTO：附作者、点赞数/点赞状态、楼中楼回复数/层数与贴纸展示信息。
func buildCommentInfo(servants *server.Servants, c *modelinteracts.Comment, momentID uint64, likes uint32, liked bool, replyCount int64, replyDepth int) (resp.CommentInfo, error) {
	author, err := servants.UserServant.QueryByID(c.UserID)
	if err != nil {
		return resp.CommentInfo{}, err
	}
	stickers, err := Sticker.ResolveContentStickers(servants, c.Content)
	if err != nil {
		return resp.CommentInfo{}, err
	}
	return toCommentInfoWith(c, momentID, *author, likes, liked, replyCount, replyDepth, stickers), nil
}

// buildReplyInfo 组装楼中楼回复响应 DTO：附回复作者、点赞数/点赞状态与贴纸展示信息。
func buildReplyInfo(servants *server.Servants, c *modelinteracts.Comment, momentID uint64, replyToUserID uint32, likes uint32, liked bool) (resp.ReplyInfo, error) {
	author, err := servants.UserServant.QueryByID(c.UserID)
	if err != nil {
		return resp.ReplyInfo{}, err
	}
	stickers, err := Sticker.ResolveContentStickers(servants, c.Content)
	if err != nil {
		return resp.ReplyInfo{}, err
	}
	return toReplyInfoWith(c, momentID, replyToUserID, *author, likes, liked, stickers), nil
}

func toCommentInfoWith(c *modelinteracts.Comment, momentID uint64, author modeluser.User, likes uint32, liked bool, replyCount int64, replyDepth int, stickers []resp.CommentSticker) resp.CommentInfo {
	return resp.CommentInfo{
		ID:         c.ID,
		UserID:     c.UserID,
		MomentID:   momentID,
		Content:    c.Content,
		Status:     c.Status,
		CreatedAt:  c.CreatedAt,
		Author:     author,
		Likes:      likes,
		IsLiked:    liked,
		ReplyCount: replyCount,
		ReplyDepth: replyDepth,
		Stickers:   stickers,
	}
}

// maxReplyDepth 计算扁平回复列表相对首条评论的最大嵌套层数（首条评论为 0 层，直接回复为 1 层）。
// 回复按 id 升序返回，而回复只能指向更早创建的评论（父节点 id 更小），因此按序一次遍历即可得到各回复层数；
// 父节点缺失（被删除等）时按第 1 层处理。
func maxReplyDepth(rootID uint64, replies []modelinteracts.Comment) int {
	depth := make(map[uint64]int, len(replies)+1)
	depth[rootID] = 0
	max := 0
	for i := range replies {
		d, ok := depth[replies[i].TargetID]
		if !ok {
			d = 0
		}
		d++
		depth[replies[i].ID] = d
		if d > max {
			max = d
		}
	}
	return max
}

func toReplyInfoWith(c *modelinteracts.Comment, momentID uint64, replyToUserID uint32, author modeluser.User, likes uint32, liked bool, stickers []resp.CommentSticker) resp.ReplyInfo {
	return resp.ReplyInfo{
		ID:            c.ID,
		UserID:        c.UserID,
		MomentID:      momentID,
		ReplyToID:     c.TargetID,
		ReplyToUserID: replyToUserID,
		Content:       c.Content,
		Status:        c.Status,
		CreatedAt:     c.CreatedAt,
		Author:        author,
		Likes:         likes,
		IsLiked:       liked,
		Stickers:      stickers,
	}
}

// checkCommentStickers 校验评论内容中的贴纸使用并按出现顺序返回贴纸 UUID 列表（无贴纸返回 nil）。
// 产品规则：登录且绑定手机号的用户可使用贴纸；一条评论最多 consts.MaxCommentStickerTokens 张贴纸；
// 仅可使用本人上传/已加入收藏夹/已收藏贴纸包内且未被封禁的贴纸。
func checkCommentStickers(ctx fiber.Ctx, servants *server.Servants, uid uint32, content string) ([]string, error) {
	stickerUUIDs, err := Sticker.CheckCommentStickers(servants, uid, content)
	if err != nil {
		switch {
		case errors.Is(err, Sticker.ErrTokenExceeded):
			return nil, resp.StickerInvalid(ctx)
		case errors.Is(err, Sticker.ErrNotFound):
			return nil, resp.StickerNotFound(ctx)
		case errors.Is(err, Sticker.ErrPackBanned):
			return nil, resp.StickerPackBanned(ctx)
		case errors.Is(err, Sticker.ErrNotUsable):
			return nil, resp.StickerNotUsable(ctx)
		default:
			return nil, resp.ServerError(ctx)
		}
	}
	if len(stickerUUIDs) == 0 {
		return nil, nil
	}

	bound, err := servants.UserServant.HasCredential(uid, consts.Phone)
	if err != nil {
		return nil, resp.ServerError(ctx)
	}
	if !bound {
		return nil, resp.PhoneNotBound(ctx)
	}
	return stickerUUIDs, nil
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
