package UserFile

import (
	"context"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
	storages3 "miaoverse/service/s3"
)

const TimeFormat = "2006-01-02 15:04:05"

func BuildObjectKey(uid uint32, fileUUID string, fileName string) string {
	return fmt.Sprintf("uploads/%d/%s/%s", uid, fileUUID, fileName)
}

func SanitizeFileName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\\", "/")
	value = filepath.Base(value)
	value = strings.Trim(value, ". ")
	if value == "" || value == "/" {
		return ""
	}
	return value
}

// FileTypeName 将文件大类 uint8 常量转为对外字符串
func FileTypeName(fileType uint8) string {
	switch fileType {
	case modeluser.FileTypeImage:
		return "image"
	case modeluser.FileTypeVideo:
		return "video"
	case modeluser.FileTypeAudio:
		return "audio"
	case modeluser.FileTypeDocument:
		return "document"
	default:
		return "other"
	}
}

// HashHex 将二进制 hash 转为 hex 字符串
func HashHex(hash [32]byte) string {
	return hex.EncodeToString(hash[:])
}

func ToFileInfo(file *modeluser.File) resp.FileInfo {
	if file == nil {
		return resp.FileInfo{}
	}
	return resp.FileInfo{
		UUID:      file.UUID,
		FileName:  file.FileName,
		FileURL:   file.FileURL,
		FileType:  FileTypeName(file.FileType),
		FileExt:   file.FileExt,
		MimeType:  file.MimeType,
		FileSize:  file.FileSize,
		Hash:      HashHex(file.Hash),
		CreatedAt: file.CreatedAt.Format(TimeFormat),
	}
}

// BuildSharedTempLink 为任意用户的 active 文件生成绑定请求者身份的临时访问链接
func BuildSharedTempLink(ctx context.Context, s3 *storages3.Servant, requesterUID uint32, file *modeluser.File) (*resp.TempFileLink, error) {
	uidValue := fmt.Sprintf("%d", requesterUID)
	signature, err := s3.CreateTempSignature(uidValue)
	if err != nil {
		return nil, err
	}
	link, err := s3.GetTempObjectLink(ctx, uidValue, signature.Signature, file.ObjectKey)
	if err != nil {
		return nil, err
	}
	return &resp.TempFileLink{
		UUID:      file.UUID,
		URL:       link.URL,
		ExpiresAt: link.ExpiresAt,
	}, nil
}
