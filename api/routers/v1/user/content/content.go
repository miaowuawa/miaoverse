package content

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/middleware"
	modelmoment "miaoverse/model/dao/moment"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
	"miaoverse/service/UserMoment"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

func CountHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	targetID, ok := middleware.BlockTarget(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	isFriend, isFan, err := relationFlags(ctx, servants, uid, targetID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	count, err := servants.ContentServant.CountVisibleMomentsByUser(uid, targetID, isFriend, isFan)
	if err != nil {
		return resp.ServerError(ctx)
	}

	return resp.ContentCount(ctx, count)
}

func ListHandler(ctx fiber.Ctx, servants *server.Servants) error {
	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	targetID, ok := middleware.BlockTarget(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	offset, limit, ok := parsePagination(ctx)
	if !ok {
		return resp.BadRequest(ctx)
	}

	isFriend, isFan, err := relationFlags(ctx, servants, uid, targetID)
	if err != nil {
		return resp.ServerError(ctx)
	}

	moments, err := servants.ContentServant.QueryVisibleMomentsByUser(uid, targetID, isFriend, isFan, offset, limit)
	if err != nil {
		return resp.ServerError(ctx)
	}

	metas, err := servants.ContentServant.QueryMomentMetas(momentIDs(moments))
	if err != nil {
		return resp.ServerError(ctx)
	}

	items := make([]resp.ContentItem, 0, len(moments))
	for i := range moments {
		items = append(items, UserMoment.ToContentItem(&moments[i], metas))
	}

	return resp.ContentList(ctx, items)
}

func parsePagination(ctx fiber.Ctx) (int, int, bool) {
	offset := 0
	limit := defaultListLimit
	if raw := strings.TrimSpace(ctx.Query("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, false
		}
		offset = value
	}
	if raw := strings.TrimSpace(ctx.Query("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > maxListLimit {
			return 0, 0, false
		}
		limit = value
	}
	return offset, limit, true
}

// relationFlags 计算查看者与目标用户的关系：isFriend 互相关注，isFan 目标关注了查看者
func relationFlags(ctx fiber.Ctx, servants *server.Servants, viewerID uint32, targetID uint32) (bool, bool, error) {
	if viewerID == targetID {
		return false, false, nil
	}

	following, err := servants.InteractsServant.IsFollowing(viewerID, targetID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, false, err
	}
	followedBy, err := servants.InteractsServant.IsFollowing(targetID, viewerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, false, err
	}
	return following && followedBy, followedBy, nil
}

func momentIDs(moments []modelmoment.Moment) []uint64 {
	ids := make([]uint64, 0, len(moments))
	for i := range moments {
		ids = append(ids, moments[i].ID)
	}
	return ids
}
