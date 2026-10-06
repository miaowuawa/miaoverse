package resp

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/service/i18n"
)

func JSON(ctx fiber.Ctx, status int, msg i18n.MessageKey) error {
	return ctx.Status(status).JSON(CodeWithMsg{
		Code: status,
		Msg:  i18n.Message(ctx, msg),
	})
}

// Blocked 拉黑/被拉黑关系导致的拒绝响应，body 中 code 为自定义业务错误码 40301
func Blocked(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusForbidden).JSON(CodeWithMsg{
		Code: consts.BlockedByRelation,
		Msg:  i18n.Message(ctx, i18n.ErrBlockedByRelation),
	})
}

// Punished 权限封禁导致的拒绝响应，body 中 code 为自定义业务错误码 40302
func Punished(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusForbidden).JSON(CodeWithMsg{
		Code: consts.Punished,
		Msg:  i18n.Message(ctx, i18n.ErrPunished),
	})
}

// AccountBanned 账号封禁（不允许登录）导致的拒绝响应，body 中 code 为自定义业务错误码 40303
func AccountBanned(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusForbidden).JSON(CodeWithMsg{
		Code: consts.AccountBanned,
		Msg:  i18n.Message(ctx, i18n.ErrAccountBanned),
	})
}

// TargetPunished 目标用户存在生效中的权限封禁，body 中 code 为自定义业务错误码 40304
func TargetPunished(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusForbidden).JSON(CodeWithMsg{
		Code: consts.TargetPunished,
		Msg:  i18n.Message(ctx, i18n.ErrTargetPunished),
	})
}

// ContentBlocked 内容被屏蔽（标记为屏蔽状态，如违规内容）导致无法查看，
// 使用 HTTP 451 Unavailable For Legal Reasons，body 中 code 为自定义业务错误码 45101。
func ContentBlocked(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusUnavailableForLegalReasons).JSON(CodeWithMsg{
		Code: consts.ContentBlocked,
		Msg:  i18n.Message(ctx, i18n.ErrContentBlocked),
	})
}

func BadRequest(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusBadRequest, i18n.ErrBadRequest)
}

func Unauthorized(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusUnauthorized, i18n.ErrUnauthorized)
}

// NeedLogin 需要登录但未登录的拒绝响应，HTTP 401，body 中 code 为自定义业务错误码 40101。
// 与 Unauthorized 的区别：40101 用于"接口本身需要登录"（如 following feed），
// 而 Unauthorized 用于登录态失效/缺失的通用场景。
func NeedLogin(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusUnauthorized).JSON(CodeWithMsg{
		Code: consts.NeedLogin,
		Msg:  i18n.Message(ctx, i18n.ErrNeedLogin),
	})
}

func ServerError(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusInternalServerError, i18n.ErrServerContactAdmin)
}

func StorageUnavailable(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusServiceUnavailable, i18n.ErrS3Unavailable)
}

func FileTooLarge(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusRequestEntityTooLarge, i18n.ErrFileTooLarge)
}

// FileImageInvalid 图片上传未通过安全图片校验（非 jpg/png/gif/webp 或伪装上传「图片藏 JS」），body 中 code 为 HTTP 400。
func FileImageInvalid(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusBadRequest, i18n.ErrFileImageInvalid)
}

func FileNotFound(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusNotFound, i18n.ErrFileNotFound)
}

func UserNotFound(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusNotFound, i18n.ErrUserNotFound)
}

func FileNotShared(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusForbidden, i18n.ErrFileNotShared)
}

func FileBlockedByOwner(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusForbidden, i18n.ErrFileBlockedByOwner)
}

func FileUploaded(ctx fiber.Ctx, file FileInfo) error {
	return ctx.Status(fiber.StatusCreated).JSON(CodeWithMsgFile{
		Code: fiber.StatusCreated,
		Msg:  i18n.Message(ctx, i18n.OKFileUploaded),
		File: file,
	})
}

func FileTempLink(ctx fiber.Ctx, fileUUID string, url string, expiresAt time.Time) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgTempFileLink{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKFileTempLink),
		Link: TempFileLink{
			UUID:      fileUUID,
			URL:       url,
			ExpiresAt: expiresAt,
		},
	})
}

// AvatarUpdated 头像设置成功响应，返回头像文件 UUID（不返回原始存储 URL）。
func AvatarUpdated(ctx fiber.Ctx, avatarUUID string) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgAvatar{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKAvatarUpdated),
		Avatar: AvatarInfo{
			AvatarUUID: avatarUUID,
		},
	})
}

// AvatarFetched 头像获取成功响应，返回头像文件 UUID（不返回原始存储 URL）。
func AvatarFetched(ctx fiber.Ctx, avatarUUID string) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgAvatar{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKAvatarFetched),
		Avatar: AvatarInfo{
			AvatarUUID: avatarUUID,
		},
	})
}

func MomentPublished(ctx fiber.Ctx, moment MomentInfo) error {
	return ctx.Status(fiber.StatusCreated).JSON(CodeWithMsgMoment{
		Code:   fiber.StatusCreated,
		Msg:    i18n.Message(ctx, i18n.OKMomentPublished),
		Moment: moment,
	})
}

func MomentUpdated(ctx fiber.Ctx, moment MomentInfo) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgMoment{
		Code:   fiber.StatusOK,
		Msg:    i18n.Message(ctx, i18n.OKMomentUpdated),
		Moment: moment,
	})
}

// MomentDetailOK 返回动态详情，已注销作者的展示字段统一打码，避免泄露注销前的用户名与签名。
func MomentDetailOK(ctx fiber.Ctx, detail MomentDetail) error {
	MaskClosedAccount(ctx, &detail.Author)
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgMomentDetail{
		Code:   fiber.StatusOK,
		Msg:    i18n.Message(ctx, i18n.OKMomentDetailFetched),
		Moment: detail,
	})
}

func BlockUpdated(ctx fiber.Ctx, target uint32, blockType uint8, action string) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgBlock{
		Code:   fiber.StatusOK,
		Msg:    i18n.Message(ctx, i18n.OKBlockUpdated),
		Target: target,
		Type:   blockType,
		Action: action,
	})
}

func ContentCount(ctx fiber.Ctx, count int64) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgContentList{
		Code:  fiber.StatusOK,
		Msg:   i18n.Message(ctx, i18n.OKContentList),
		Count: count,
	})
}

func ContentList(ctx fiber.Ctx, contents []ContentItem) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgContentList{
		Code:     fiber.StatusOK,
		Msg:      i18n.Message(ctx, i18n.OKContentList),
		Contents: contents,
	})
}

// FeedList 返回 feed 列表，作者已注销时展示字段统一打码。
func FeedList(ctx fiber.Ctx, count int64, items []FeedItem) error {
	for i := range items {
		MaskClosedAccount(ctx, &items[i].Author)
	}
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgFeedList{
		Code:  fiber.StatusOK,
		Msg:   i18n.Message(ctx, i18n.OKFeedFetched),
		Count: count,
		Items: items,
	})
}

func UserInfoOK(ctx fiber.Ctx, info UserInfo) error {
	MaskClosedAccount(ctx, &info.User)
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgUserInfo{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKUserInfoFetched),
		User: info,
	})
}

func CommentCreated(ctx fiber.Ctx, comment CommentInfo) error {
	MaskClosedAccount(ctx, &comment.Author)
	return ctx.Status(fiber.StatusCreated).JSON(CodeWithMsgComment{
		Code:    fiber.StatusCreated,
		Msg:     i18n.Message(ctx, i18n.OKCommentCreated),
		Comment: comment,
	})
}

// CommentList 返回动态一级评论分页列表，已注销作者的展示字段统一打码。
func CommentList(ctx fiber.Ctx, count int64, comments []CommentInfo) error {
	for i := range comments {
		MaskClosedAccount(ctx, &comments[i].Author)
	}
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgCommentList{
		Code:     fiber.StatusOK,
		Msg:      i18n.Message(ctx, i18n.OKCommentsFetched),
		Count:    count,
		Comments: comments,
	})
}

// ReplyCreated 返回新发表的楼中楼回复，已注销作者的展示字段统一打码。
func ReplyCreated(ctx fiber.Ctx, reply ReplyInfo) error {
	MaskClosedAccount(ctx, &reply.Author)
	return ctx.Status(fiber.StatusCreated).JSON(CodeWithMsgReply{
		Code:  fiber.StatusCreated,
		Msg:   i18n.Message(ctx, i18n.OKReplyCreated),
		Reply: reply,
	})
}

// Conversation 返回楼中楼完整对话，已注销作者（首条评论与回复）的展示字段统一打码。
func Conversation(ctx fiber.Ctx, conversation ConversationInfo) error {
	MaskClosedAccount(ctx, &conversation.Root.Author)
	for i := range conversation.Replies {
		MaskClosedAccount(ctx, &conversation.Replies[i].Author)
	}
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgConversation{
		Code:         fiber.StatusOK,
		Msg:          i18n.Message(ctx, i18n.OKConversationFetched),
		Conversation: conversation,
	})
}

func RelationList(ctx fiber.Ctx, count int64, users []RelationUser) error {
	for i := range users {
		MaskClosedAccount(ctx, &users[i].User)
	}
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgRelationList{
		Code:  fiber.StatusOK,
		Msg:   i18n.Message(ctx, i18n.OKRelationList),
		Count: count,
		Users: users,
	})
}

func InteractOK(ctx fiber.Ctx, target uint64, action string) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgInteract{
		Code:   fiber.StatusOK,
		Msg:    i18n.Message(ctx, i18n.OKInteract),
		Target: target,
		Action: action,
	})
}

// ===== 贴纸 =====

func StickerUploaded(ctx fiber.Ctx, sticker StickerInfo) error {
	return ctx.Status(fiber.StatusCreated).JSON(CodeWithMsgSticker{
		Code:    fiber.StatusCreated,
		Msg:     i18n.Message(ctx, i18n.OKStickerUploaded),
		Sticker: sticker,
	})
}

func StickerOK(ctx fiber.Ctx, sticker StickerInfo) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgSticker{
		Code:    fiber.StatusOK,
		Msg:     i18n.Message(ctx, i18n.OKStickerUpdated),
		Sticker: sticker,
	})
}

func StickerList(ctx fiber.Ctx, count int64, stickers []StickerInfo) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgStickerList{
		Code:     fiber.StatusOK,
		Msg:      i18n.Message(ctx, i18n.OKStickerList),
		Count:    count,
		Stickers: stickers,
	})
}

func StickerPackCreated(ctx fiber.Ctx, pack StickerPackInfo) error {
	return ctx.Status(fiber.StatusCreated).JSON(CodeWithMsgStickerPack{
		Code: fiber.StatusCreated,
		Msg:  i18n.Message(ctx, i18n.OKStickerPackCreated),
		Pack: pack,
	})
}

func StickerPackOK(ctx fiber.Ctx, pack StickerPackInfo) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgStickerPack{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKStickerUpdated),
		Pack: pack,
	})
}

func StickerPackList(ctx fiber.Ctx, count int64, packs []StickerPackInfo) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgStickerPackList{
		Code:  fiber.StatusOK,
		Msg:   i18n.Message(ctx, i18n.OKStickerPackList),
		Count: count,
		Packs: packs,
	})
}

func StickerPackDetail(ctx fiber.Ctx, pack StickerPackInfo, stickers []StickerInfo) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgStickerPackDetail{
		Code:     fiber.StatusOK,
		Msg:      i18n.Message(ctx, i18n.OKStickerList),
		Pack:     pack,
		Stickers: stickers,
	})
}

func StickerNotFound(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusNotFound, i18n.ErrStickerNotFound)
}

func StickerPackNotFound(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusNotFound, i18n.ErrStickerPackNotFound)
}

func StickerTooLarge(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusRequestEntityTooLarge, i18n.ErrStickerTooLarge)
}

func StickerImageInvalid(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusBadRequest, i18n.ErrStickerImageInvalid)
}

func StickerInvalid(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusBadRequest, i18n.ErrStickerInvalid)
}

func StickerNotUsable(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusForbidden, i18n.ErrStickerNotUsable)
}

func StickerPackBanned(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusForbidden, i18n.ErrStickerPackBanned)
}

func StickerPackFull(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusBadRequest, i18n.ErrStickerPackFull)
}

func StickerFavoritesFull(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusBadRequest, i18n.ErrStickerFavoritesFull)
}

// PhoneNotBound 需要绑定手机号才能执行的操作（如上传/使用贴纸），body 中 code 为 HTTP 403。
func PhoneNotBound(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusForbidden, i18n.ErrPhoneNotBound)
}
