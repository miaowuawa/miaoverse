package relation

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/middleware"
	modelinteracts "miaoverse/model/dao/interacts"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

func FollowingHandler(ctx fiber.Ctx, servants *server.Servants) error {
	return listRelation(ctx, servants, false)
}

func FollowersHandler(ctx fiber.Ctx, servants *server.Servants) error {
	return listRelation(ctx, servants, true)
}

// listRelation 分页查询目标用户的关注列表（isFollower=false）或粉丝列表（isFollower=true）。
func listRelation(ctx fiber.Ctx, servants *server.Servants, isFollower bool) error {
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

	var (
		interacts []modelinteracts.Interacts
		count     int64
		err       error
	)
	if isFollower {
		count, err = servants.InteractsServant.CountFollowers(targetID)
		if err != nil {
			return resp.ServerError(ctx)
		}
		interacts, err = servants.InteractsServant.QueryFollowersByUser(targetID, offset, limit)
	} else {
		count, err = servants.InteractsServant.CountFollowing(targetID)
		if err != nil {
			return resp.ServerError(ctx)
		}
		interacts, err = servants.InteractsServant.QueryFollowingByUser(targetID, offset, limit)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			interacts = nil
		} else {
			return resp.ServerError(ctx)
		}
	}

	ids := make([]uint32, 0, len(interacts))
	for i := range interacts {
		if isFollower {
			ids = append(ids, interacts[i].UserFrom)
		} else {
			ids = append(ids, interacts[i].UserTo)
		}
	}

	users, err := servants.UserServant.QueryUsersByIDs(ids)
	if err != nil {
		return resp.ServerError(ctx)
	}

	items := make([]resp.RelationUser, 0, len(interacts))
	for i := range interacts {
		var userID uint32
		if isFollower {
			userID = interacts[i].UserFrom
		} else {
			userID = interacts[i].UserTo
		}
		u, exists := users[userID]
		if !exists {
			continue
		}
		blockStatus, err := servants.BlockServant.GetBlockStatus(ctx.Context(), uid, userID)
		if err != nil {
			return resp.ServerError(ctx)
		}
		items = append(items, resp.RelationUser{
			User:        u,
			BlockStatus: blockStatus,
		})
	}

	return resp.RelationList(ctx, count, items)
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
