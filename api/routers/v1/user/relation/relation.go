package relation

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/consts"
	"miaoverse/middleware"
	modelinteracts "miaoverse/model/dao/interacts"
	"miaoverse/model/dto/resp"
	"miaoverse/model/server"
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

	viewerFollows, targetFollowsViewer, err := servants.InteractsServant.QueryFollowStatusBatch(uid, ids)
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
			User:         u,
			BlockStatus:  blockStatus,
			FollowStatus: followStatus(viewerFollows[userID], targetFollowsViewer[userID]),
		})
	}

	return resp.RelationList(ctx, count, items)
}

// followStatus 以当前查看者为视角计算关注关系状态。
func followStatus(viewerFollows, targetFollowsViewer bool) uint8 {
	switch {
	case viewerFollows && targetFollowsViewer:
		return consts.FollowStatusMutual
	case viewerFollows:
		return consts.FollowStatusFollowing
	case targetFollowsViewer:
		return consts.FollowStatusFollowedBy
	default:
		return consts.FollowStatusNone
	}
}

func parsePagination(ctx fiber.Ctx) (int, int, bool) {
	offset := 0
	limit := consts.DefaultListLimit
	if raw := strings.TrimSpace(ctx.Query("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, false
		}
		offset = value
	}
	if raw := strings.TrimSpace(ctx.Query("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > consts.MaxListLimit {
			return 0, 0, false
		}
		limit = value
	}
	return offset, limit, true
}
