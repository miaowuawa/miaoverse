package file

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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
	"miaoverse/service/UserFile"
	"miaoverse/util"
)

const formFileField = "file"

func UploadHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if servants.S3Servant == nil {
		return resp.StorageUnavailable(ctx)
	}

	fileHeader, err := ctx.FormFile(formFileField)
	if err != nil || fileHeader == nil {
		return resp.BadRequest(ctx)
	}
	if servants.MaxUploadFileSize > 0 && fileHeader.Size > servants.MaxUploadFileSize {
		return resp.FileTooLarge(ctx)
	}

	req := uploadreq.UploadFile{
		FileType: strings.TrimSpace(ctx.FormValue("file_type")),
	}
	fileName := UserFile.SanitizeFileName(fileHeader.Filename)
	if fileName == "" {
		return resp.BadRequest(ctx)
	}

	fileHash, err := hashUploadedFile(fileHeader)
	if err != nil {
		return resp.ServerError(ctx)
	}

	fileUUID := uuid.NewString()
	mimeType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	fileExt := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	fileType := util.FileType.Normalize(req.FileType, mimeType)

	reusedFile, err := servants.UserServant.QueryActiveFileByHash(fileHash)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return resp.ServerError(ctx)
	}

	recordInput := modeluser.File{
		UUID:     fileUUID,
		UserID:   uid,
		FileName: fileName,
		FileType: fileType,
		FileExt:  fileExt,
		MimeType: mimeType,
		FileSize: uint64(fileHeader.Size),
		Hash:     fileHash,
		Status:   modeluser.FileStatusActive,
	}

	if reusedFile != nil && reusedFile.ID != 0 {
		recordInput.ObjectKey = reusedFile.ObjectKey
		recordInput.FileURL = reusedFile.FileURL
		record, err := servants.UserServant.CreateFile(recordInput)
		if err != nil {
			return resp.ServerError(ctx)
		}
		return resp.FileUploaded(ctx, UserFile.ToFileInfo(record))
	}

	objectKey := UserFile.BuildObjectKey(uid, fileUUID, fileName)
	src, err := fileHeader.Open()
	if err != nil {
		return resp.ServerError(ctx)
	}
	defer src.Close()

	fileURL, err := servants.S3Servant.PutObject(ctx.Context(), objectKey, src, mimeType)
	if err != nil {
		return resp.ServerError(ctx)
	}

	recordInput.ObjectKey = objectKey
	recordInput.FileURL = fileURL
	record, err := servants.UserServant.CreateFile(recordInput)
	if err != nil {
		_ = servants.S3Servant.DeleteObject(ctx.Context(), objectKey)
		return resp.ServerError(ctx)
	}

	return resp.FileUploaded(ctx, UserFile.ToFileInfo(record))
}

func TempLinkHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if servants.S3Servant == nil {
		return resp.StorageUnavailable(ctx)
	}

	fileUUID := strings.TrimSpace(ctx.Params("uuid"))
	if fileUUID == "" {
		return resp.BadRequest(ctx)
	}
	if _, err := uuid.Parse(fileUUID); err != nil {
		return resp.BadRequest(ctx)
	}

	record, err := servants.UserServant.QueryActiveFileByUUID(uid, fileUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.FileNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}

	uidValue := fmt.Sprintf("%d", uid)
	signature, err := servants.S3Servant.CreateTempSignature(uidValue)
	if err != nil {
		return resp.ServerError(ctx)
	}
	link, err := servants.S3Servant.GetTempObjectLink(ctx.Context(), uidValue, signature.Signature, record.ObjectKey)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.FileTempLink(ctx, record.UUID, link.URL, link.ExpiresAt)
}

func hashUploadedFile(fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
