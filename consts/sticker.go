package consts

// 贴纸域常量：状态、来源、封禁标记、限制。

const (
	StickerStatusActive  uint8 = 1 // 正常
	StickerStatusDeleted uint8 = 4 // 删除
)

// 贴纸收藏夹条目来源（user_sticker.source）。
const (
	StickerSourceOwn      uint8 = 1 // 本人上传（上传时自动加入收藏夹）
	StickerSourceFavorite uint8 = 2 // 收藏他人贴纸
)

// 贴纸包封禁标记（sticker_pack.banned）。
// 贴纸包被封禁后，包内贴纸在所有使用处（评论等）无法显示。
const (
	StickerPackBannedNone uint8 = 0 // 正常
	StickerPackBannedFlag uint8 = 1 // 已封禁
)

// 贴纸包状态（sticker_pack.status）。
const (
	StickerPackStatusActive  uint8 = 1 // 正常
	StickerPackStatusDeleted uint8 = 4 // 删除
)

const (
	StickerTopNone   uint8 = 0 // 不置顶
	StickerTopPinned uint8 = 1 // 置顶（收藏夹内优先展示）
)

const (
	DefaultStickerMaxFileSizeBytes int64 = 10 * 1024 * 1024 // 单张贴纸最大字节数（默认 10MB，可配置覆盖）
	MaxStickerFavorites                  = 500              // 个人收藏夹贴纸数量上限（含收藏他人贴纸）
	MaxStickerPackItems                  = 100              // 单个贴纸包贴纸数量上限
	MaxCommentStickerTokens              = 25               // 一条评论最多使用的贴纸数
	MaxStickerNameLen                    = 64               // 贴纸名称最大长度
	MaxStickerPackNameLen                = 64               // 贴纸包名称最大长度
	MaxStickerPackDescLen                = 255              // 贴纸包描述最大长度
)
