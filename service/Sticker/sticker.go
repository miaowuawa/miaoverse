package Sticker

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"miaoverse/consts"
	modelsticker "miaoverse/model/dao/sticker"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
)

// 贴纸在评论文本中的内嵌标记：[sticker:<uuid>]。
// 标记由前端贴纸选择器在光标处插入，作为评论文字的一部分穿插展示；
// 一条评论最多使用一个贴纸（consts.MaxCommentStickerTokens）。
const (
	tokenPrefix = "[sticker:"
	tokenSuffix = "]"
)

// tokenRegexp 只接受合法 UUID 形式的贴纸标记，普通文本中的其他方括号内容不会被误判。
var tokenRegexp = regexp.MustCompile(`\[sticker:([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})\]`)

// 贴纸使用校验的业务错误（由 handler 统一映射为 resp 错误响应）。
var (
	ErrTokenExceeded = errors.New("sticker tokens exceeded") // 一条评论使用了多个贴纸
	ErrNotFound      = errors.New("sticker not found")       // 贴纸不存在或已删除
	ErrPackBanned    = errors.New("sticker pack banned")     // 贴纸所在贴纸包被封禁
	ErrNotUsable     = errors.New("sticker not usable")      // 非本人上传/未收藏，无权使用
)

// BuildToken 生成贴纸内嵌标记。
func BuildToken(stickerUUID string) string {
	return tokenPrefix + stickerUUID + tokenSuffix
}

// ExtractTokens 按出现顺序提取文本中的全部贴纸 UUID。
func ExtractTokens(content string) []string {
	matches := tokenRegexp.FindAllStringSubmatch(content, -1)
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		result = append(result, strings.ToLower(m[1]))
	}
	return result
}

// CheckCommentSticker 校验评论内容中的贴纸使用并返回贴纸 UUID（无贴纸时返回空串）：
//   - 内嵌标记数不得超过 consts.MaxCommentStickerTokens（一条评论最多一个贴纸）；
//   - 贴纸必须存在且未删除；
//   - 贴纸必须是使用者本人上传、已加入收藏夹、或属于已收藏贴纸包（登录且绑定手机号用户的可用集合）；
//   - 所在贴纸包被封禁的贴纸不可用于新评论。
func CheckCommentSticker(servants *server.Servants, uid uint32, content string) (string, error) {
	tokens := ExtractTokens(content)
	if len(tokens) == 0 {
		return "", nil
	}
	if len(tokens) > consts.MaxCommentStickerTokens {
		return "", ErrTokenExceeded
	}

	stickerUUID := tokens[0]
	s, err := servants.StickerServant.QueryStickerByUUID(stickerUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}

	banned, err := IsPackBanned(servants, s)
	if err != nil {
		return "", err
	}
	if banned {
		return "", ErrPackBanned
	}

	usable, err := IsUsable(servants, uid, s)
	if err != nil {
		return "", err
	}
	if !usable {
		return "", ErrNotUsable
	}

	return stickerUUID, nil
}

// IsUsable 判断用户是否可以使用该贴纸：本人上传、已加入个人收藏夹、或属于已收藏贴纸包内的贴纸。
func IsUsable(servants *server.Servants, uid uint32, s *modelsticker.Sticker) (bool, error) {
	if s.UserID == uid {
		return true, nil
	}

	own, err := servants.StickerServant.QueryUserSticker(uid, s.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if own != nil && own.ID != 0 {
		return true, nil
	}

	if s.PackID == 0 {
		return false, nil
	}
	favorited, err := servants.StickerServant.QueryFavoritedPackStickerIDsBatch(uid, []uint64{s.ID})
	if err != nil {
		return false, err
	}
	return favorited[s.ID], nil
}

// IsPackBanned 判断贴纸所在贴纸包是否被封禁（无贴纸包的贴纸恒为 false）。
// 贴纸包不存在或已删除时按封禁处理（包内贴纸一律不可用/不可显示）。
func IsPackBanned(servants *server.Servants, s *modelsticker.Sticker) (bool, error) {
	if s.PackID == 0 {
		return false, nil
	}
	pack, err := servants.StickerServant.QueryPackByID(s.PackID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}
	return pack.Banned != consts.StickerPackBannedNone, nil
}

// NormalizeName 清洗贴纸/贴纸包名称：去首尾空白并按最大长度截断（按 rune 截断，避免截断多字节字符）。
func NormalizeName(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > maxLen {
		runes = runes[:maxLen]
		value = string(runes)
	}
	return value
}

// ToStickerInfo DAO 贴纸 → 响应 DTO。source/top 为收藏夹视角字段（不在收藏夹时为 0）。
func ToStickerInfo(s *modelsticker.Sticker, source uint8, top uint8) resp.StickerInfo {
	if s == nil {
		return resp.StickerInfo{}
	}
	return resp.StickerInfo{
		UUID:      s.UUID,
		FileUUID:  s.FileUUID,
		PackID:    s.PackID,
		Name:      s.Name,
		Source:    source,
		Top:       top,
		CreatedAt: s.CreatedAt.Format(consts.TimeFormat),
	}
}

// ToPackInfo DAO 贴纸包 → 响应 DTO。
func ToPackInfo(p *modelsticker.Pack, stickerCount int64, isFavorite bool) resp.StickerPackInfo {
	if p == nil {
		return resp.StickerPackInfo{}
	}
	return resp.StickerPackInfo{
		ID:           p.ID,
		UUID:         p.UUID,
		UserID:       p.UserID,
		Name:         p.Name,
		Description:  p.Description,
		Banned:       p.Banned != consts.StickerPackBannedNone,
		StickerCount: stickerCount,
		IsFavorite:   isFavorite,
		CreatedAt:    p.CreatedAt.Format(consts.TimeFormat),
	}
}

// ResolveCommentStickers 批量解析评论内嵌贴纸的展示信息：
// 贴纸不存在/已删除或所在贴纸包被封禁时 Hidden=true（前端隐藏贴纸并在评论下提示「部分贴纸未显示」）。
// 返回 uuid → 展示信息；传入的 uuid 都会出现在结果中。
func ResolveCommentStickers(servants *server.Servants, stickerUUIDs []string) (map[string]resp.CommentSticker, error) {
	result := map[string]resp.CommentSticker{}
	unique := make([]string, 0, len(stickerUUIDs))
	seen := map[string]bool{}
	for _, u := range stickerUUIDs {
		u = strings.ToLower(strings.TrimSpace(u))
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		unique = append(unique, u)
		result[u] = resp.CommentSticker{UUID: u, Hidden: true}
	}
	if len(unique) == 0 {
		return result, nil
	}

	stickers, err := servants.StickerServant.QueryStickersByUUIDsBatch(unique)
	if err != nil {
		return nil, err
	}

	packIDs := make([]uint64, 0, len(stickers))
	for _, s := range stickers {
		if s.PackID != 0 {
			packIDs = append(packIDs, s.PackID)
		}
	}
	packs, err := servants.StickerServant.QueryPacksByIDsBatch(packIDs)
	if err != nil {
		return nil, err
	}

	for _, u := range unique {
		s, ok := stickers[u]
		if !ok {
			continue
		}
		hidden := false
		if s.PackID != 0 {
			pack, ok := packs[s.PackID]
			if !ok || pack.Banned != consts.StickerPackBannedNone {
				hidden = true
			}
		}
		result[u] = resp.CommentSticker{
			UUID:     s.UUID,
			FileUUID: s.FileUUID,
			Name:     s.Name,
			Hidden:   hidden,
		}
	}
	return result, nil
}

// ValidateUUID 校验贴纸/文件 UUID 字符串格式。
func ValidateUUID(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", false
	}
	if _, err := uuid.Parse(value); err != nil {
		return "", false
	}
	return value, true
}
