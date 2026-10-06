package Sticker

import (
	"context"
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"miaoverse/consts"
	modelsticker "miaoverse/model/dao/sticker"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/server"
	"miaoverse/service/UserFile"
)

// 贴纸上传的业务错误（由 handler 统一映射为 resp 错误响应）。
var (
	ErrFileTooLarge       = errors.New("sticker file too large")
	ErrImageInvalid       = errors.New("sticker image invalid")
	ErrNameInvalid        = errors.New("sticker name invalid")
	ErrStorageUnavailable = errors.New("s3 storage unavailable")
)

// safeImageExts 安全栅格图片 MIME → 存储扩展名。
// 存储文件名固定为 sticker.<ext>（不含任何用户输入），扩展名只由嗅探出的真实类型决定。
var safeImageExts = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/webp": "webp",
}

// Upload 上传贴纸图片：
//   - 大小不超过 maxSize 字节（配置 upload.max_sticker_size_bytes，默认 10MB）；
//   - 安全图片校验：按文件头魔数嗅探真实类型，仅接受 jpg/png/gif/webp，拒绝 SVG/HTML/JS 等
//     可携带脚本的格式以及「声明 image/png 实际为 SVG」的伪装上传（防上传图片藏 JS 的存储型 XSS）；
//   - 复用 files 表与 S3 存储（文件 permission=0 公开，评论中所有可见用户可展示），相同 hash 复用已有对象不重复上传；
//   - 创建贴纸记录并自动加入上传者贴纸收藏夹（source=own）。
//
// name 为可选的贴纸展示名，缺省使用原文件名（不含扩展名）。
func Upload(ctx context.Context, servants *server.Servants, uid uint32, fileHeader *multipart.FileHeader, name string, maxSize int64) (*modelsticker.Sticker, error) {
	if fileHeader == nil {
		return nil, ErrImageInvalid
	}
	if maxSize > 0 && fileHeader.Size > maxSize {
		return nil, ErrFileTooLarge
	}
	fileName := UserFile.SanitizeFileName(fileHeader.Filename)
	if fileName == "" {
		return nil, ErrNameInvalid
	}
	if servants.S3Servant == nil {
		return nil, ErrStorageUnavailable
	}

	// 安全：只信文件头魔数嗅探出的真实类型，并要求声明 MIME（如有）同在白名单内（防「图片藏 JS」）
	detectedMIME, safe, err := UserFile.DetectSafeUploadedImage(fileHeader, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	if !safe {
		return nil, ErrImageInvalid
	}
	ext := safeImageExts[detectedMIME]

	displayName := NormalizeName(name, consts.MaxStickerNameLen)
	if displayName == "" {
		displayName = NormalizeName(strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName)), consts.MaxStickerNameLen)
	}

	fileHash, err := UserFile.HashUploadedFile(fileHeader)
	if err != nil {
		return nil, err
	}

	fileUUID := uuid.NewString()
	recordInput := modeluser.File{
		UUID:       fileUUID,
		UserID:     uid,
		FileName:   fileName,
		FileType:   consts.FileTypeImage,
		FileExt:    ext,
		MimeType:   detectedMIME,
		FileSize:   uint64(fileHeader.Size),
		Permission: consts.FilePermissionPublic,
		Hash:       fileHash,
		Status:     consts.FileStatusActive,
	}

	// 相同 hash 文件复用已有对象，不重复上传到 S3
	reusedFile, err := servants.UserServant.QueryActiveFileByHash(fileHash)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	uploadedKey := ""
	if reusedFile != nil && reusedFile.ID != 0 {
		recordInput.ObjectKey = reusedFile.ObjectKey
		recordInput.FileURL = reusedFile.FileURL
	} else {
		// 存储文件名不含任何用户输入，扩展名由嗅探出的真实类型决定
		objectKey := UserFile.BuildObjectKey(uid, fileUUID, "sticker."+ext)
		src, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()

		fileURL, err := servants.S3Servant.PutObject(ctx, objectKey, src, detectedMIME)
		if err != nil {
			return nil, err
		}
		recordInput.ObjectKey = objectKey
		recordInput.FileURL = fileURL
		uploadedKey = objectKey
	}

	if _, err := servants.UserServant.CreateFile(recordInput); err != nil {
		if uploadedKey != "" {
			_ = servants.S3Servant.DeleteObject(ctx, uploadedKey)
		}
		return nil, err
	}

	created, err := servants.StickerServant.CreateSticker(modelsticker.Sticker{
		UUID:     uuid.NewString(),
		UserID:   uid,
		FileUUID: fileUUID,
		Name:     displayName,
		Status:   consts.StickerStatusActive,
	})
	if err != nil {
		if uploadedKey != "" {
			_ = servants.S3Servant.DeleteObject(ctx, uploadedKey)
		}
		return nil, err
	}

	// 上传的贴纸自动加入本人贴纸收藏夹（source=own），在贴纸选择器「收藏夹」中展示
	if _, err := servants.StickerServant.CreateUserSticker(modelsticker.UserSticker{
		UserID:    uid,
		StickerID: created.ID,
		Source:    consts.StickerSourceOwn,
		Top:       consts.StickerTopNone,
	}); err != nil {
		return nil, err
	}

	return created, nil
}
