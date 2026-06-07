package file

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"miaoverse/middleware"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/file/uploadreq"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/i18n"
)

const (
	formFileField = "file"
	timeFormat    = "2006-01-02 15:04:05"
)

func UploadHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return unauthorized(ctx)
	}
	if servants.S3Servant == nil {
		return storageUnavailable(ctx)
	}

	fileHeader, err := ctx.FormFile(formFileField)
	if err != nil || fileHeader == nil {
		return badRequest(ctx)
	}

	maxSize := servants.MaxUploadFileSize
	if maxSize > 0 && fileHeader.Size > maxSize {
		return ctx.Status(fiber.StatusRequestEntityTooLarge).JSON(resp.CodeWithMsg{
			Code: fiber.StatusRequestEntityTooLarge,
			Msg:  i18n.Message(ctx, i18n.ErrFileTooLarge),
		})
	}

	req := uploadreq.UploadFile{
		FileType: strings.TrimSpace(ctx.FormValue("file_type")),
	}
	fileName := sanitizeFileName(fileHeader.Filename)
	if fileName == "" {
		return badRequest(ctx)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return serverError(ctx)
	}
	defer src.Close()

	fileUUID := uuid.NewString()
	mimeType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	fileExt := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	fileType := normalizeFileType(req.FileType, mimeType)
	objectKey := buildObjectKey(uid, fileUUID, fileName)

	hasher := sha256.New()
	fileURL, err := servants.S3Servant.PutObject(ctx.Context(), objectKey, io.TeeReader(src, hasher), mimeType)
	if err != nil {
		return serverError(ctx)
	}

	record, err := servants.UserServant.CreateFile(modeluser.File{
		UUID:      fileUUID,
		UserID:    uid,
		FileName:  fileName,
		ObjectKey: objectKey,
		FileURL:   fileURL,
		FileType:  fileType,
		FileExt:   fileExt,
		MimeType:  mimeType,
		FileSize:  uint64(fileHeader.Size),
		Hash:      hex.EncodeToString(hasher.Sum(nil)),
		Status:    modeluser.FileStatusActive,
	})
	if err != nil {
		_ = servants.S3Servant.DeleteObject(ctx.Context(), objectKey)
		return serverError(ctx)
	}

	return ctx.Status(fiber.StatusCreated).JSON(resp.CodeWithMsgFile{
		Code: fiber.StatusCreated,
		Msg:  i18n.Message(ctx, i18n.OKFileUploaded),
		File: toFileInfo(record),
	})
}

func TempLinkHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return unauthorized(ctx)
	}
	if servants.S3Servant == nil {
		return storageUnavailable(ctx)
	}

	fileUUID := strings.TrimSpace(ctx.Params("uuid"))
	if fileUUID == "" {
		return badRequest(ctx)
	}
	if _, err := uuid.Parse(fileUUID); err != nil {
		return badRequest(ctx)
	}

	record, err := servants.UserServant.QueryActiveFileByUUID(uid, fileUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(resp.CodeWithMsg{
				Code: fiber.StatusNotFound,
				Msg:  i18n.Message(ctx, i18n.ErrFileNotFound),
			})
		}
		return serverError(ctx)
	}

	signature, err := servants.S3Servant.CreateTempSignature(fmt.Sprintf("%d", uid))
	if err != nil {
		return serverError(ctx)
	}
	link, err := servants.S3Servant.GetTempObjectLink(ctx.Context(), fmt.Sprintf("%d", uid), signature.Signature, record.ObjectKey)
	if err != nil {
		return serverError(ctx)
	}

	return ctx.Status(fiber.StatusOK).JSON(resp.CodeWithMsgTempFileLink{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKFileTempLink),
		Link: resp.TempFileLink{
			UUID:      record.UUID,
			URL:       link.URL,
			ExpiresAt: link.ExpiresAt,
		},
	})
}

func buildObjectKey(uid uint64, fileUUID string, fileName string) string {
	return fmt.Sprintf("uploads/%d/%s/%s", uid, fileUUID, fileName)
}

func sanitizeFileName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\\", "/")
	value = filepath.Base(value)
	value = strings.Trim(value, ". ")
	if value == "" || value == "/" {
		return ""
	}
	return value
}

func normalizeFileType(value string, mimeType string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "image", "video", "audio", "document", "other":
		return value
	}

	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(mimeType))
	}
	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return "image"
	case strings.HasPrefix(mediaType, "video/"):
		return "video"
	case strings.HasPrefix(mediaType, "audio/"):
		return "audio"
	case mediaType == "application/pdf", strings.HasPrefix(mediaType, "text/"), strings.Contains(mediaType, "document"):
		return "document"
	default:
		return "other"
	}
}

func toFileInfo(file *modeluser.File) resp.FileInfo {
	if file == nil {
		return resp.FileInfo{}
	}
	return resp.FileInfo{
		UUID:      file.UUID,
		FileName:  file.FileName,
		FileURL:   file.FileURL,
		FileType:  file.FileType,
		FileExt:   file.FileExt,
		MimeType:  file.MimeType,
		FileSize:  file.FileSize,
		Hash:      file.Hash,
		CreatedAt: file.CreatedAt.Format(timeFormat),
	}
}

func badRequest(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(resp.CodeWithMsg{
		Code: fiber.StatusBadRequest,
		Msg:  i18n.Message(ctx, i18n.ErrBadRequest),
	})
}

func unauthorized(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusUnauthorized).JSON(resp.CodeWithMsg{
		Code: fiber.StatusUnauthorized,
		Msg:  i18n.Message(ctx, i18n.ErrUnauthorized),
	})
}

func storageUnavailable(ctx fiber.Ctx) error {
	return ctx.Status(fiber.StatusServiceUnavailable).JSON(resp.CodeWithMsg{
		Code: fiber.StatusServiceUnavailable,
		Msg:  i18n.Message(ctx, i18n.ErrS3Unavailable),
	})
}

func serverError(ctx fiber.Ctx) error {
	return ctx.Status(http.StatusInternalServerError).JSON(resp.CodeWithMsg{
		Code: http.StatusInternalServerError,
		Msg:  i18n.Message(ctx, i18n.ErrServerContactAdmin),
	})
}
