package file

import (
	"crypto/sha256"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"miaoverse/consts"
	"miaoverse/middleware"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/file/uploadreq"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/UserFile"
	"miaoverse/util"
)

func UploadHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if servants.S3Servant == nil {
		return resp.StorageUnavailable(ctx)
	}

	fileHeader, err := ctx.FormFile(consts.FormFileField)
	if err != nil || fileHeader == nil {
		return resp.BadRequest(ctx)
	}
	if servants.MaxUploadFileSize > 0 && fileHeader.Size > servants.MaxUploadFileSize {
		return resp.FileTooLarge(ctx)
	}

	req := uploadreq.UploadFile{
		FileType: strings.TrimSpace(ctx.FormValue("file_type")),
	}
	permission := req.Permission
	if permission != consts.FilePermissionPublic &&
		permission != consts.FilePermissionFriends &&
		permission != consts.FilePermissionFans &&
		permission != consts.FilePermissionNone {
		return resp.BadRequest(ctx)
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
		UUID:       fileUUID,
		UserID:     uid,
		FileName:   fileName,
		FileType:   fileType,
		FileExt:    fileExt,
		MimeType:   mimeType,
		FileSize:   uint64(fileHeader.Size),
		Permission: permission,
		Hash:       fileHash,
		Status:     consts.FileStatusActive,
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

	link, err := UserFile.BuildSharedTempLink(ctx.Context(), servants.S3Servant, uid, record)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.FileTempLink(ctx, record.UUID, link.URL, link.ExpiresAt)
}

// SharedTempLinkHandler 获取任意用户 active 文件的临时访问链接，用于帖子等查看/下载他人文件/媒体
// 未登录用户可访问公开（permission=0）文件；登录用户按访问控制规则判定。
func SharedTempLinkHandler(ctx fiber.Ctx, servants *server.Servants) error {
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

	record, err := servants.UserServant.QueryActiveFileByUUIDAnyUser(fileUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.FileNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}

	uid, loggedIn := middleware.CurrentUID(ctx)
	if !loggedIn {
		// 匿名访问：仅允许公开（permission=0）文件，其余一律按不存在处理，避免泄露文件存在性；
		// 临时链接使用固定匿名身份标识签名，不绑定任何用户
		if record.Permission != consts.FilePermissionPublic {
			return resp.FileNotFound(ctx)
		}
		link, err := UserFile.BuildAnonymousSharedTempLink(ctx.Context(), servants.S3Servant, record)
		if err != nil {
			return resp.ServerError(ctx)
		}
		return resp.FileTempLink(ctx, record.UUID, link.URL, link.ExpiresAt)
	}

	if err := UserFile.CheckSharedAccess(ctx.Context(), servants.BlockServant, servants.InteractsServant, uid, record.UserID, record.Permission); err != nil {
		if errors.Is(err, UserFile.ErrFileBlockedByOwner) {
			return resp.FileBlockedByOwner(ctx)
		}
		if errors.Is(err, UserFile.ErrFileNotShared) {
			return resp.FileNotShared(ctx)
		}
		return resp.ServerError(ctx)
	}

	link, err := UserFile.BuildSharedTempLink(ctx.Context(), servants.S3Servant, uid, record)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.FileTempLink(ctx, record.UUID, link.URL, link.ExpiresAt)
}

func hashUploadedFile(fileHeader *multipart.FileHeader) ([32]byte, error) {
	var fileHash [32]byte
	src, err := fileHeader.Open()
	if err != nil {
		return fileHash, err
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return fileHash, err
	}
	copy(fileHash[:], hasher.Sum(nil))
	return fileHash, nil
}
