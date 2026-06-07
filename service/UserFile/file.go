package UserFile

import (
	"fmt"
	"path/filepath"
	"strings"

	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
)

const TimeFormat = "2006-01-02 15:04:05"

func BuildObjectKey(uid uint64, fileUUID string, fileName string) string {
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

func ToFileInfo(file *modeluser.File) resp.FileInfo {
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
		CreatedAt: file.CreatedAt.Format(TimeFormat),
	}
}
