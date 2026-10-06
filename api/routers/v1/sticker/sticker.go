package sticker

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/consts"
	"miaoverse/middleware"
	modelsticker "miaoverse/model/dao/sticker"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/sticker/stickerreq"
	"miaoverse/model/server"
	"miaoverse/service/Sticker"
	"miaoverse/util/pagination"
)

// UploadHandler 上传贴纸（multipart/form-data：file 必填，name 可选展示名）。
// 登录、绑定手机号、未被封禁上传权限由路由中间件校验；单张大小限制从配置读取（默认 10MB）。
// 安全：只接受 jpg/png/gif/webp 安全栅格图片（按文件头魔数嗅探），拒绝 SVG/HTML/JS 等可携带脚本的格式。
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

	name := ctx.FormValue("name")
	created, err := Sticker.Upload(ctx.Context(), servants, uid, fileHeader, name, servants.MaxStickerFileSize)
	if err != nil {
		switch {
		case errors.Is(err, Sticker.ErrFileTooLarge):
			return resp.StickerTooLarge(ctx)
		case errors.Is(err, Sticker.ErrImageInvalid):
			return resp.StickerImageInvalid(ctx)
		case errors.Is(err, Sticker.ErrNameInvalid):
			return resp.BadRequest(ctx)
		case errors.Is(err, Sticker.ErrStorageUnavailable):
			return resp.StorageUnavailable(ctx)
		default:
			return resp.ServerError(ctx)
		}
	}

	source, top, err := collectionMeta(servants, uid, created.ID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.StickerUploaded(ctx, Sticker.ToStickerInfo(created, source, top))
}

// CollectionHandler 获取当前用户贴纸收藏夹（我上传的 + 收藏的贴纸，置顶优先）。
// 个人收藏夹最多添加 500 张收藏贴纸（consts.MaxStickerFavorites）。
func CollectionHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	offset, limit, ok := pagination.Parse(ctx.Query("offset"), ctx.Query("limit"))
	if !ok {
		return resp.BadRequest(ctx)
	}

	rows, err := servants.StickerServant.QueryUserStickers(uid, offset, limit)
	if err != nil {
		return resp.ServerError(ctx)
	}
	count, err := servants.StickerServant.CountUserStickers(uid, 0)
	if err != nil {
		return resp.ServerError(ctx)
	}

	items, err := toStickerInfoList(servants, rows)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.StickerList(ctx, count, items)
}

// FavoriteHandler 把他人贴纸添加到当前用户贴纸收藏夹（幂等）。
// 校验贴纸存在、未被封禁；收藏数量达到上限（500）时拒绝。
func FavoriteHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &stickerreq.FavoriteSticker{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	stickerUUID, ok := Sticker.ValidateUUID(req.StickerUUID)
	if !ok {
		return resp.BadRequest(ctx)
	}

	s, err := servants.StickerServant.QueryStickerByUUID(stickerUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.StickerNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}
	banned, err := Sticker.IsPackBanned(servants, s)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if banned {
		return resp.StickerPackBanned(ctx)
	}

	// 已在收藏夹（含本人上传条目）时幂等返回
	if existing, err := servants.StickerServant.QueryUserSticker(uid, s.ID); err == nil && existing != nil && existing.ID != 0 {
		return resp.StickerOK(ctx, Sticker.ToStickerInfo(s, existing.Source, existing.Top))
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return resp.ServerError(ctx)
	}

	// 上限只约束「收藏他人贴纸」（source=2）；本人上传的贴纸自动入收藏夹不占该额度
	favorites, err := servants.StickerServant.CountUserStickers(uid, consts.StickerSourceFavorite)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if favorites >= consts.MaxStickerFavorites {
		return resp.StickerFavoritesFull(ctx)
	}

	created, err := servants.StickerServant.CreateUserSticker(modelsticker.UserSticker{
		UserID:    uid,
		StickerID: s.ID,
		Source:    consts.StickerSourceFavorite,
		Top:       consts.StickerTopNone,
	})
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.StickerOK(ctx, Sticker.ToStickerInfo(s, created.Source, created.Top))
}

// UnfavoriteHandler 从收藏夹移除收藏的贴纸（幂等）。
// 本人上传的贴纸条目不在此删除（使用 DELETE /stickers/:uuid 删除贴纸本身）。
func UnfavoriteHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &stickerreq.FavoriteSticker{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	stickerUUID, ok := Sticker.ValidateUUID(req.StickerUUID)
	if !ok {
		return resp.BadRequest(ctx)
	}

	s, err := servants.StickerServant.QueryStickerByUUID(stickerUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.StickerNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}

	existing, err := servants.StickerServant.QueryUserSticker(uid, s.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return resp.ServerError(ctx)
	}
	if existing != nil && existing.ID != 0 && existing.Source == consts.StickerSourceFavorite {
		if err := servants.StickerServant.DeleteUserSticker(uid, s.ID); err != nil {
			return resp.ServerError(ctx)
		}
	}
	return resp.StickerOK(ctx, Sticker.ToStickerInfo(s, 0, consts.StickerTopNone))
}

// TopHandler 设置/取消收藏夹贴纸置顶（body: top 0|1）。贴纸必须在当前用户收藏夹中。
func TopHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &stickerreq.SetStickerTop{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	if req.Top != consts.StickerTopNone && req.Top != consts.StickerTopPinned {
		return resp.BadRequest(ctx)
	}

	s, err := queryOwnSticker(ctx, servants, uid)
	if err != nil {
		return err
	}
	existing, err := servants.StickerServant.QueryUserSticker(uid, s.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.StickerNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}
	if err := servants.StickerServant.UpdateUserStickerTop(uid, s.ID, req.Top); err != nil {
		return resp.ServerError(ctx)
	}
	return resp.StickerOK(ctx, Sticker.ToStickerInfo(s, existing.Source, req.Top))
}

// DeleteHandler 删除本人上传的贴纸（软删除，同时清理收藏夹条目）。
// 贴纸包内引用随贴纸状态置为删除，历史评论中的贴纸按「贴纸未显示」处理。
func DeleteHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	s, err := queryOwnSticker(ctx, servants, uid)
	if err != nil {
		return err
	}
	if err := servants.StickerServant.DeleteSticker(s.ID); err != nil {
		return resp.ServerError(ctx)
	}
	if err := servants.StickerServant.DeleteUserSticker(uid, s.ID); err != nil {
		return resp.ServerError(ctx)
	}
	return resp.InteractOK(ctx, s.ID, consts.ActionRemove)
}

// queryOwnSticker 按路径参数 :uuid 查询属于当前用户的 active 贴纸。
func queryOwnSticker(ctx fiber.Ctx, servants *server.Servants, uid uint32) (*modelsticker.Sticker, error) {
	stickerUUID, ok := Sticker.ValidateUUID(ctx.Params("uuid"))
	if !ok {
		return nil, resp.BadRequest(ctx)
	}
	s, err := servants.StickerServant.QueryStickerByUUID(stickerUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, resp.StickerNotFound(ctx)
		}
		return nil, resp.ServerError(ctx)
	}
	if s.UserID != uid {
		return nil, resp.StickerNotFound(ctx)
	}
	return s, nil
}

// collectionMeta 查询贴纸在当前用户收藏夹中的来源与置顶状态（不在收藏夹时为 0）。
func collectionMeta(servants *server.Servants, uid uint32, stickerID uint64) (uint8, uint8, error) {
	us, err := servants.StickerServant.QueryUserSticker(uid, stickerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, consts.StickerTopNone, nil
		}
		return 0, 0, err
	}
	return us.Source, us.Top, nil
}

// toStickerInfoList 收藏夹条目 → 贴纸响应列表（贴纸已删除的条目跳过）。
func toStickerInfoList(servants *server.Servants, rows []modelsticker.UserSticker) ([]resp.StickerInfo, error) {
	ids := make([]uint64, 0, len(rows))
	for i := range rows {
		ids = append(ids, rows[i].StickerID)
	}
	stickers, err := servants.StickerServant.QueryStickersByIDsBatch(ids)
	if err != nil {
		return nil, err
	}
	items := make([]resp.StickerInfo, 0, len(rows))
	for i := range rows {
		s, ok := stickers[rows[i].StickerID]
		if !ok {
			continue
		}
		items = append(items, Sticker.ToStickerInfo(&s, rows[i].Source, rows[i].Top))
	}
	return items, nil
}
