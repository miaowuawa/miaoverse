package UserFile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"miaoverse/consts"
	"miaoverse/dao/interacts"
	modeluser "miaoverse/model/dao/user"
	"miaoverse/model/dto/resp"
	"miaoverse/service/UserBlock"
	storages3 "miaoverse/service/s3"
	"miaoverse/util/filetype"
)

var (
	ErrFileNotShared      = errors.New("file is not shared")
	ErrFileBlockedByOwner = errors.New("file blocked by owner")

	// AnonymousTempLinkUID 匿名临时链接使用的固定身份标识（不指向任何真实用户）
	AnonymousTempLinkUID = "anonymous"
)

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

// HashUploadedFile 计算上传文件的 SHA-256 hash（相同 hash 文件复用存储，避免重复上传）。
func HashUploadedFile(fileHeader *multipart.FileHeader) ([32]byte, error) {
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

// ImageHeadSniffBytes 图片魔数嗅探读取的文件头字节数。
const ImageHeadSniffBytes = 512

// DetectSafeUploadedImage 检测上传文件是否为安全栅格图片（jpg/png/gif/webp），返回嗅探出的真实 MIME。
// 用于图片上传防「图片藏 JS」（存储型 XSS）：
//   - 按文件头魔数嗅探真实类型，SVG/HTML/JS/XML 等可携带脚本的格式一律拒绝；
//   - 声明 MIME 与真实内容不一致（如声明 image/png 实际为 SVG）同样拒绝；
//   - ok=false 表示不通过；读取文件头失败时 err 非 nil。
//
// 调用方应以返回的嗅探 MIME 作为存储 mime_type 与 S3 Content-Type，避免浏览器按伪造类型解析。
func DetectSafeUploadedImage(fileHeader *multipart.FileHeader, declaredMIME string) (string, bool, error) {
	head, err := ReadHead(fileHeader, ImageHeadSniffBytes)
	if err != nil {
		return "", false, err
	}
	mimeType, ok := filetype.DetectSafeImageMIME(head, declaredMIME)
	return mimeType, ok, nil
}

// ReadHead 读取上传文件头（最多 n 字节），用于魔数嗅探。
func ReadHead(fileHeader *multipart.FileHeader, n int) ([]byte, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	head := make([]byte, n)
	read, err := io.ReadFull(src, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return head[:read], nil
}

// FileTypeName 将文件大类 uint8 常量转为对外字符串
func FileTypeName(fileType uint8) string {
	switch fileType {
	case consts.FileTypeImage:
		return "image"
	case consts.FileTypeVideo:
		return "video"
	case consts.FileTypeAudio:
		return "audio"
	case consts.FileTypeDocument:
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
		CreatedAt: file.CreatedAt.Format(consts.TimeFormat),
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

// BuildAnonymousSharedTempLink 为公开文件生成不绑定任何用户的临时访问链接（未登录查看场景）。
// 使用固定匿名标识做签名，不涉及用户身份信息。
func BuildAnonymousSharedTempLink(ctx context.Context, s3 *storages3.Servant, file *modeluser.File) (*resp.TempFileLink, error) {
	signature, err := s3.CreateTempSignature(AnonymousTempLinkUID)
	if err != nil {
		return nil, err
	}
	link, err := s3.GetTempObjectLink(ctx, AnonymousTempLinkUID, signature.Signature, file.ObjectKey)
	if err != nil {
		return nil, err
	}
	return &resp.TempFileLink{
		UUID:      file.UUID,
		URL:       link.URL,
		ExpiresAt: link.ExpiresAt,
	}, nil
}

// CheckSharedAccess 校验查看者是否有权访问他人文件。
// 被查看的用户拉黑了查看者时，无论文件是否公开都拒绝；
// 否则按文件分享权限（公开/好友/粉丝/不公开）判定。
func CheckSharedAccess(ctx context.Context, block *UserBlock.Servant, interactsServant *interacts.InteractsDAO, requesterUID uint32, ownerUID uint32, permission uint8) error {
	if requesterUID == ownerUID {
		return nil
	}

	blocked, err := block.Contains(ctx, ownerUID, UserBlock.BlockTypeBlock, requesterUID)
	if err != nil {
		return err
	}
	if blocked {
		return ErrFileBlockedByOwner
	}

	switch permission {
	case consts.FilePermissionPublic:
		return nil
	case consts.FilePermissionFriends:
		following, err := interactsServant.IsFollowing(requesterUID, ownerUID)
		if err != nil {
			return err
		}
		if !following {
			return ErrFileNotShared
		}
		return nil
	case consts.FilePermissionFans:
		following, err := interactsServant.IsFollowing(ownerUID, requesterUID)
		if err != nil {
			return err
		}
		if !following {
			return ErrFileNotShared
		}
		return nil
	default:
		return ErrFileNotShared
	}
}
