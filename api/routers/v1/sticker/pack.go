package sticker

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
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

// PackCreateHandler 创建贴纸包（登录、绑定手机号、未被封禁评论权限由路由中间件校验）。
func PackCreateHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &stickerreq.CreateStickerPack{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	// 名称/描述按 rune 截断到上限（避免截断多字节字符），名称不能为空
	name := Sticker.NormalizeName(req.Name, consts.MaxStickerPackNameLen)
	if name == "" {
		return resp.BadRequest(ctx)
	}
	description := Sticker.NormalizeName(req.Description, consts.MaxStickerPackDescLen)

	created, err := servants.StickerServant.CreatePack(modelsticker.Pack{
		UUID:        uuid.NewString(),
		UserID:      uid,
		Name:        name,
		Description: description,
		Status:      consts.StickerPackStatusActive,
	})
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.StickerPackCreated(ctx, Sticker.ToPackInfo(created, 0, false))
}

// PackListHandler 贴纸包列表（未封禁）。
// filter=all（默认）全部贴纸包；filter=favorite 当前用户已收藏的贴纸包（收藏的整包内容随包更新自动同步）。
func PackListHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	offset, limit, ok := pagination.Parse(ctx.Query("offset"), ctx.Query("limit"))
	if !ok {
		return resp.BadRequest(ctx)
	}

	filter := strings.TrimSpace(ctx.Query("filter"))
	if filter == "" {
		filter = "all"
	}

	var (
		packs []modelsticker.Pack
		count int64
		err   error
	)
	switch filter {
	case "all":
		packs, err = servants.StickerServant.QueryPacks(offset, limit)
		if err != nil {
			return resp.ServerError(ctx)
		}
		count, err = servants.StickerServant.CountPacks()
	case "favorite":
		packs, err = servants.StickerServant.QueryFavoritePacks(uid, offset, limit)
		if err != nil {
			return resp.ServerError(ctx)
		}
		count, err = servants.StickerServant.CountFavoritePacks(uid)
	default:
		return resp.BadRequest(ctx)
	}
	if err != nil {
		return resp.ServerError(ctx)
	}

	items, err := toPackInfoList(servants, uid, packs)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.StickerPackList(ctx, count, items)
}

// PackDetailHandler 贴纸包详情（含包内贴纸列表）。被封禁的贴纸包返回 403。
func PackDetailHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	pack, err := queryPackByPathID(ctx, servants)
	if err != nil {
		return err
	}
	if pack.Banned != consts.StickerPackBannedNone {
		return resp.StickerPackBanned(ctx)
	}

	stickers, err := servants.StickerServant.QueryStickersByPack(pack.ID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	favorites, err := servants.StickerServant.QueryFavoritePackIDs(uid)
	if err != nil {
		return resp.ServerError(ctx)
	}

	items := make([]resp.StickerInfo, 0, len(stickers))
	for i := range stickers {
		items = append(items, Sticker.ToStickerInfo(&stickers[i], 0, consts.StickerTopNone))
	}
	return resp.StickerPackDetail(ctx, Sticker.ToPackInfo(pack, int64(len(items)), favorites[pack.ID]), items)
}

// PackAddStickerHandler 把本人上传的贴纸加入自己的贴纸包（单包上限 100 张）。
func PackAddStickerHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &stickerreq.AddStickerToPack{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}
	stickerUUID, ok := Sticker.ValidateUUID(req.StickerUUID)
	if !ok {
		return resp.BadRequest(ctx)
	}

	pack, err := queryOwnedPackByPathID(ctx, servants, uid)
	if err != nil {
		return err
	}

	s, err := servants.StickerServant.QueryStickerByUUID(stickerUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.StickerNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}
	// 只能打包本人上传的贴纸；已在其他贴纸包的贴纸先移出
	if s.UserID != uid {
		return resp.StickerNotUsable(ctx)
	}
	if s.PackID != 0 && s.PackID != pack.ID {
		return resp.BadRequest(ctx)
	}
	if s.PackID == pack.ID {
		return resp.StickerOK(ctx, Sticker.ToStickerInfo(s, 0, consts.StickerTopNone))
	}

	count, err := servants.StickerServant.CountStickersByPack(pack.ID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	if count >= consts.MaxStickerPackItems {
		return resp.StickerPackFull(ctx)
	}

	if err := servants.StickerServant.UpdateStickerPack(s.ID, pack.ID, uint32(count+1)); err != nil {
		return resp.ServerError(ctx)
	}
	s.PackID = pack.ID
	return resp.StickerOK(ctx, Sticker.ToStickerInfo(s, 0, consts.StickerTopNone))
}

// PackRemoveStickerHandler 把贴纸移出自己的贴纸包（贴纸本身保留）。
func PackRemoveStickerHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	pack, err := queryOwnedPackByPathID(ctx, servants, uid)
	if err != nil {
		return err
	}

	stickerUUID, ok := Sticker.ValidateUUID(ctx.Params("sticker_uuid"))
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
	if s.PackID != pack.ID {
		return resp.StickerNotFound(ctx)
	}

	if err := servants.StickerServant.UpdateStickerPack(s.ID, 0, 0); err != nil {
		return resp.ServerError(ctx)
	}
	s.PackID = 0
	return resp.StickerOK(ctx, Sticker.ToStickerInfo(s, 0, consts.StickerTopNone))
}

// PackFavoriteHandler 收藏整个贴纸包（幂等）。收藏后包内容更新自动同步展示。
func PackFavoriteHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &stickerreq.FavoriteStickerPack{}
	if err := ctx.Bind().Body(req); err != nil || req.PackID == 0 {
		return resp.BadRequest(ctx)
	}

	pack, err := servants.StickerServant.QueryPackByID(req.PackID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.StickerPackNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}
	if pack.Banned != consts.StickerPackBannedNone {
		return resp.StickerPackBanned(ctx)
	}

	if _, err := servants.StickerServant.QueryPackFavorite(uid, pack.ID); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.ServerError(ctx)
		}
		if err := servants.StickerServant.CreatePackFavorite(uid, pack.ID); err != nil {
			return resp.ServerError(ctx)
		}
	}

	count, err := servants.StickerServant.CountStickersByPack(pack.ID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.StickerPackOK(ctx, Sticker.ToPackInfo(pack, count, true))
}

// PackUnfavoriteHandler 取消收藏贴纸包（幂等）。
func PackUnfavoriteHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	req := &stickerreq.FavoriteStickerPack{}
	if err := ctx.Bind().Body(req); err != nil || req.PackID == 0 {
		return resp.BadRequest(ctx)
	}

	pack, err := servants.StickerServant.QueryPackByID(req.PackID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp.StickerPackNotFound(ctx)
		}
		return resp.ServerError(ctx)
	}
	if err := servants.StickerServant.DeletePackFavorite(uid, pack.ID); err != nil {
		return resp.ServerError(ctx)
	}

	count, err := servants.StickerServant.CountStickersByPack(pack.ID)
	if err != nil {
		return resp.ServerError(ctx)
	}
	return resp.StickerPackOK(ctx, Sticker.ToPackInfo(pack, count, false))
}

// queryPackByPathID 按路径参数 :id 查询 active 贴纸包。
func queryPackByPathID(ctx fiber.Ctx, servants *server.Servants) (*modelsticker.Pack, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(ctx.Params("id")), 10, 64)
	if err != nil || id == 0 {
		return nil, resp.BadRequest(ctx)
	}
	pack, err := servants.StickerServant.QueryPackByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, resp.StickerPackNotFound(ctx)
		}
		return nil, resp.ServerError(ctx)
	}
	return pack, nil
}

// queryOwnedPackByPathID 按路径参数 :id 查询属于当前用户的 active 贴纸包（非本人一律按不存在处理）。
func queryOwnedPackByPathID(ctx fiber.Ctx, servants *server.Servants, uid uint32) (*modelsticker.Pack, error) {
	pack, err := queryPackByPathID(ctx, servants)
	if err != nil {
		return nil, err
	}
	if pack.UserID != uid {
		return nil, resp.StickerPackNotFound(ctx)
	}
	return pack, nil
}

// toPackInfoList 贴纸包列表 → 响应列表（附包内贴纸数与收藏状态）。
func toPackInfoList(servants *server.Servants, uid uint32, packs []modelsticker.Pack) ([]resp.StickerPackInfo, error) {
	favorites, err := servants.StickerServant.QueryFavoritePackIDs(uid)
	if err != nil {
		return nil, err
	}
	items := make([]resp.StickerPackInfo, 0, len(packs))
	for i := range packs {
		count, err := servants.StickerServant.CountStickersByPack(packs[i].ID)
		if err != nil {
			return nil, err
		}
		items = append(items, Sticker.ToPackInfo(&packs[i], count, favorites[packs[i].ID]))
	}
	return items, nil
}
