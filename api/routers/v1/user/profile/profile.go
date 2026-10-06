package profile

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"miaoverse/middleware"
	"miaoverse/model/dto/resp"
	"miaoverse/model/dto/user/updatereq"
	"miaoverse/model/server"
	"miaoverse/service/UserProfile"
	"miaoverse/service/i18n"
)

// UpdateFullHandler 全量更新本人基础资料（账号名、昵称、头像、个性签名、性别）。
// 用户身份一律取自登录会话，请求体中的用户标识字段不会被采纳，因此不存在越权改他人资料的路径。
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

	fields, ok := fullFields(req)
	if !ok {
		return resp.BadRequest(ctx)
	}
	// 头像涉及文件归属校验，无法放进纯函数校验里，单独追加
	if req.Avatar != nil {
		field, err := UserProfile.AvatarField(servants.UserServant, uid, *req.Avatar)
		if err != nil {
			return avatarError(ctx, err)
		}
		fields = append(fields, field)
	}
	return updateProfile(ctx, servants, uid, fields)
}

// UpdatePartialHandler 部分更新本人基础资料，只有显式传入的字段会被修改。
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

	fields, ok := patchFields(req)
	if !ok {
		return resp.BadRequest(ctx)
	}
	// 头像涉及文件归属校验，无法放进纯函数校验里，单独追加
	if req.Avatar != nil {
		field, err := UserProfile.AvatarField(servants.UserServant, uid, *req.Avatar)
		if err != nil {
			return avatarError(ctx, err)
		}
		fields = append(fields, field)
	}
	if len(fields) == 0 {
		return resp.BadRequest(ctx)
	}
	return updateProfile(ctx, servants, uid, fields)
}

// avatarError 把头像校验错误映射为响应：
//   - ErrAvatarEmpty / ErrAvatarInvalid：400（文件不存在、非本人、非图片、非公开）
//   - 其他：500
//
// PermAvatar 权限封禁由 UserProfile.Update 统一返回 ErrPunished，由 updateProfile 映射为 40302。
func avatarError(ctx fiber.Ctx, err error) error {
	if errors.Is(err, UserProfile.ErrAvatarEmpty) || errors.Is(err, UserProfile.ErrAvatarInvalid) {
		return resp.BadRequest(ctx)
	}
	return resp.ServerError(ctx)
}

func updateProfile(ctx fiber.Ctx, servants *server.Servants, uid uint32, fields []UserProfile.Field) error {
	user, err := UserProfile.Update(servants.UserServant, uid, fields, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, UserProfile.ErrNoFields):
			return resp.BadRequest(ctx)
		case errors.Is(err, UserProfile.ErrPunished):
			// 命中生效中的权限封禁（如 PermSignature / PermNickname），返回 403 + 40302
			return resp.Punished(ctx)
		case errors.Is(err, gorm.ErrRecordNotFound):
			return resp.UserNotFound(ctx)
		case isConflict(err):
			return ctx.Status(fiber.StatusConflict).JSON(resp.CodeWithMsg{
				Code: fiber.StatusConflict,
				Msg:  i18n.Message(ctx, i18n.ErrUserInfoConflict),
			})
		default:
			return resp.ServerError(ctx)
		}
	}

	resp.MaskClosedAccount(ctx, user)
	return ctx.Status(fiber.StatusOK).JSON(resp.CodeWithMsgUser{
		Code:  fiber.StatusOK,
		Msg:   i18n.Message(ctx, i18n.OKUserInfoUpdated),
		User:  *user,
		Phone: MaskedSessionPhone(ctx),
	})
}

// fullFields 校验并转换全量更新请求。PUT 语义要求四个字段都可落库，任一字段非法即整体拒绝。
func fullFields(req *updatereq.ProfileFull) ([]UserProfile.Field, bool) {
	username, ok := UserProfile.NormalizeUsername(req.Username)
	if !ok {
		return nil, false
	}
	nickname, ok := UserProfile.NormalizeNickname(req.Nickname)
	if !ok {
		return nil, false
	}
	bio, ok := UserProfile.NormalizeBio(req.Bio)
	if !ok {
		return nil, false
	}
	gender, ok := UserProfile.NormalizeGender(req.Gender)
	if !ok {
		return nil, false
	}

	return []UserProfile.Field{
		{Column: "username", Value: username, Perm: UserProfile.PermUsername},
		{Column: "nickname", Value: nickname, Perm: UserProfile.PermNickname},
		{Column: "bio", Value: bio, Perm: UserProfile.PermBio},
		{Column: "gender", Value: gender},
	}, true
}

// patchFields 校验并转换部分更新请求，未传字段（nil）不参与更新。
// 返回的 ok 只表示「已传入的字段全部合法」，不表示「至少有一个字段」：
// 头像需要额外的文件校验，是否为空由调用方在合并头像字段后统一判断。
func patchFields(req *updatereq.ProfilePatch) ([]UserProfile.Field, bool) {
	fields := make([]UserProfile.Field, 0, 5)

	if req.Username != nil {
		username, ok := UserProfile.NormalizeUsername(*req.Username)
		if !ok {
			return nil, false
		}
		fields = append(fields, UserProfile.Field{Column: "username", Value: username, Perm: UserProfile.PermUsername})
	}
	if req.Nickname != nil {
		nickname, ok := UserProfile.NormalizeNickname(*req.Nickname)
		if !ok {
			return nil, false
		}
		fields = append(fields, UserProfile.Field{Column: "nickname", Value: nickname, Perm: UserProfile.PermNickname})
	}
	if req.Bio != nil {
		bio, ok := UserProfile.NormalizeBio(*req.Bio)
		if !ok {
			return nil, false
		}
		fields = append(fields, UserProfile.Field{Column: "bio", Value: bio, Perm: UserProfile.PermBio})
	}
	if req.Gender != nil {
		gender, ok := UserProfile.NormalizeGender(*req.Gender)
		if !ok {
			return nil, false
		}
		fields = append(fields, UserProfile.Field{Column: "gender", Value: gender})
	}

	return fields, true
}

// isConflict 判断数据库唯一键冲突（账号名、昵称各自唯一）。
// 只做错误文本匹配，任何库内异常都会继续走通用 500 分支。
func isConflict(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Duplicate entry") || strings.Contains(msg, "UNIQUE constraint failed")
}
