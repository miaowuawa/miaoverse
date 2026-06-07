package profile

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/user/updatereq"
	"miaoverse/model/server"
	"miaoverse/service/i18n"
)

const (
	maxUsernameLen = 64
	maxNicknameLen = 64
	maxAvatarLen   = 255
)

func UpdateFullHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	req := &updatereq.ProfileFull{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}

	updates, ok := fullUpdates(req)
	if !ok {
		return resp.BadRequest(ctx)
	}
	return updateProfile(ctx, servants, uid, updates)
}

func UpdatePartialHandler(ctx fiber.Ctx, servants *server.Servants) error {
	if !ctx.IsJSON() {
		return resp.BadRequest(ctx)
	}

	uid, ok := middleware.CurrentUID(ctx)
	if !ok {
		return resp.Unauthorized(ctx)
	}

	req := &updatereq.ProfilePatch{}
	if err := ctx.Bind().Body(req); err != nil {
		return resp.BadRequest(ctx)
	}

	updates, ok := patchUpdates(req)
	if !ok {
		return resp.BadRequest(ctx)
	}
	return updateProfile(ctx, servants, uid, updates)
}

func updateProfile(ctx fiber.Ctx, servants *server.Servants, uid uint64, updates map[string]any) error {
	user, err := servants.UserServant.UpdateProfile(uid, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(resp.CodeWithMsg{
				Code: fiber.StatusNotFound,
				Msg:  i18n.Message(ctx, i18n.ErrUserNotFound),
			})
		}
		if isConflict(err) {
			return ctx.Status(fiber.StatusConflict).JSON(resp.CodeWithMsg{
				Code: fiber.StatusConflict,
				Msg:  i18n.Message(ctx, i18n.ErrUserInfoConflict),
			})
		}
		return resp.ServerError(ctx)
	}

	return ctx.Status(fiber.StatusOK).JSON(resp.CodeWithMsgUser{
		Code: fiber.StatusOK,
		Msg:  i18n.Message(ctx, i18n.OKUserInfoUpdated),
		User: *user,
	})
}

func fullUpdates(req *updatereq.ProfileFull) (map[string]any, bool) {
	username := strings.TrimSpace(req.Username)
	nickname := strings.TrimSpace(req.Nickname)
	avatar := strings.TrimSpace(req.Avatar)
	if !validUsername(username) || !validNickname(nickname) || !validRegion(req.Region) || !validAvatar(avatar) || !validGender(req.Gender) {
		return nil, false
	}

	return map[string]any{
		"username": username,
		"nickname": nickname,
		"region":   req.Region,
		"avatar":   avatar,
		"gender":   req.Gender,
	}, true
}

func patchUpdates(req *updatereq.ProfilePatch) (map[string]any, bool) {
	updates := map[string]any{}

	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if !validUsername(username) {
			return nil, false
		}
		updates["username"] = username
	}
	if req.Nickname != nil {
		nickname := strings.TrimSpace(*req.Nickname)
		if !validNickname(nickname) {
			return nil, false
		}
		updates["nickname"] = nickname
	}
	if req.Region != nil {
		if !validRegion(*req.Region) {
			return nil, false
		}
		updates["region"] = *req.Region
	}
	if req.Avatar != nil {
		avatar := strings.TrimSpace(*req.Avatar)
		if !validAvatar(avatar) {
			return nil, false
		}
		updates["avatar"] = avatar
	}
	if req.Gender != nil {
		if !validGender(*req.Gender) {
			return nil, false
		}
		updates["gender"] = *req.Gender
	}

	return updates, len(updates) > 0
}

func validUsername(value string) bool {
	return value != "" && len(value) <= maxUsernameLen
}

func validNickname(value string) bool {
	return value != "" && len(value) <= maxNicknameLen
}

func validRegion(value int) bool {
	return value > 0
}

func validAvatar(value string) bool {
	return len(value) <= maxAvatarLen
}

func validGender(value uint8) bool {
	return value <= 3
}

func isConflict(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "UNIQUE constraint failed")
}
