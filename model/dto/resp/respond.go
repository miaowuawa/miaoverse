package resp

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"miaoverse/service/i18n"
)

func JSON(ctx fiber.Ctx, status int, msg i18n.MessageKey) error {
	return ctx.Status(status).JSON(CodeWithMsg{
		Code: status,
		Msg:  i18n.Message(ctx, msg),
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
