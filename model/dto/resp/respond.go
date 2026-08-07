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

func BadRequest(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusBadRequest, i18n.ErrBadRequest)
}

func Unauthorized(ctx fiber.Ctx) error {
	return JSON(ctx, fiber.StatusUnauthorized, i18n.ErrUnauthorized)
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

func MomentPublished(ctx fiber.Ctx, moment MomentInfo) error {
	return ctx.Status(fiber.StatusCreated).JSON(CodeWithMsgMoment{
		Code:   fiber.StatusCreated,
		Msg:    i18n.Message(ctx, i18n.OKMomentPublished),
		Moment: moment,
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

func UserInfoOK(ctx fiber.Ctx, info UserInfo) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgUserInfo{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKUserInfoFetched),
		User: info,
	})
}

func CommentCreated(ctx fiber.Ctx, comment CommentInfo) error {
	return ctx.Status(fiber.StatusCreated).JSON(CodeWithMsgComment{
		Code:    fiber.StatusCreated,
		Msg:     i18n.Message(ctx, i18n.OKCommentCreated),
		Comment: comment,
	})
}

func RelationList(ctx fiber.Ctx, count int64, users []RelationUser) error {
	return ctx.Status(fiber.StatusOK).JSON(CodeWithMsgRelationList{
		Code:  fiber.StatusOK,
		Msg:   i18n.Message(ctx, i18n.OKRelationList),
		Count: count,
		Users: users,
	})
}
