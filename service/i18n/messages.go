package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/pelletier/go-toml/v2"
)

type MessageKey string

const (
	LangZhCN = "zh-CN"

	ErrBadRequest            MessageKey = "error.bad_request"
	ErrRequestTimeout        MessageKey = "error.request_timeout"
	ErrServerInternal        MessageKey = "error.server_internal"
	ErrServerContactAdmin    MessageKey = "error.server_contact_admin"
	ErrSMSProvider           MessageKey = "error.sms_provider"
	ErrSMSCodeInvalid        MessageKey = "error.sms_code_invalid"
	ErrPhoneHasNoAccount     MessageKey = "error.phone_has_no_account"
	ErrNoPendingLoginAccount MessageKey = "error.no_pending_login_account"
	ErrAccountNotBelongPhone MessageKey = "error.account_not_belong_phone"
	ErrUnauthorized          MessageKey = "error.unauthorized"
	ErrUserNotFound          MessageKey = "error.user_not_found"
	ErrUserInfoConflict      MessageKey = "error.user_info_conflict"
	ErrAccountBanned         MessageKey = "error.account_banned"
	ErrAccountClosed         MessageKey = "error.account_closed"
	ErrAccountDisabled       MessageKey = "error.account_disabled"
	ErrAccountUnavailable    MessageKey = "error.account_unavailable"
	ErrPhoneNotBound         MessageKey = "error.phone_not_bound"
	ErrPasswordNotSet        MessageKey = "error.password_not_set"
	ErrCertificationRequired MessageKey = "error.certification_required"
	ErrInvalidReferrer       MessageKey = "error.invalid_referrer"
	ErrFileTooLarge          MessageKey = "error.file_too_large"
	ErrFileNotFound          MessageKey = "error.file_not_found"
	ErrFileNotShared         MessageKey = "error.file_not_shared"
	ErrFileBlockedByOwner    MessageKey = "error.file_blocked_by_owner"
	ErrBlockedByRelation     MessageKey = "error.blocked_by_relation"
	ErrPunished              MessageKey = "error.punished"
	ErrTargetPunished        MessageKey = "error.target_punished"
	ErrContentBlocked        MessageKey = "error.content_blocked"
	ErrS3Unavailable         MessageKey = "error.s3_unavailable"
	ErrNeedLogin             MessageKey = "error.need_login"

	UserClosedUsername MessageKey = "user.closed_username"
	UserClosedBio      MessageKey = "user.closed_bio"

	OKSMSSent               MessageKey = "ok.sms_sent"
	OKLogin                 MessageKey = "ok.login"
	OKRegisterAndLogin      MessageKey = "ok.register_and_login"
	OKNewAccountAndLogin    MessageKey = "ok.new_account_and_login"
	OKChooseLoginAccount    MessageKey = "ok.choose_login_account"
	OKAccountList           MessageKey = "ok.account_list"
	OKLogout                MessageKey = "ok.logout"
	OKUserInfoUpdated       MessageKey = "ok.user_info_updated"
	OKAvatarUpdated         MessageKey = "ok.avatar_updated"
	OKAvatarFetched         MessageKey = "ok.avatar_fetched"
	OKFileUploaded          MessageKey = "ok.file_uploaded"
	OKFileTempLink          MessageKey = "ok.file_temp_link"
	OKMomentPublished       MessageKey = "ok.moment_published"
	OKMomentUpdated         MessageKey = "ok.moment_updated"
	OKMomentDetailFetched   MessageKey = "ok.moment_detail_fetched"
	OKBlockUpdated          MessageKey = "ok.block_updated"
	OKContentList           MessageKey = "ok.content_list"
	OKUserInfoFetched       MessageKey = "ok.user_info_fetched"
	OKCommentCreated        MessageKey = "ok.comment_created"
	OKReplyCreated          MessageKey = "ok.reply_created"
	OKConversationFetched   MessageKey = "ok.conversation_fetched"
	OKPunishmentsFetched    MessageKey = "ok.punishments_fetched"
	OKRelationList          MessageKey = "ok.relation_list"
	OKInteract              MessageKey = "ok.interact"
	OKArticleFetched        MessageKey = "ok.article_fetched"
	OKArticleFetchedPartial MessageKey = "ok.article_fetched_partial"
	OKArticleNeedSegments   MessageKey = "ok.article_need_segments"
	OKArticleSegmentFetched MessageKey = "ok.article_segment_fetched"
	OKFeedFetched           MessageKey = "ok.feed_fetched"
	SMSActionLoginRegister  MessageKey = "sms.action.login_register"
)

type Data map[string]any

type message struct {
	Zero  string `toml:"zero"`
	One   string `toml:"one"`
	Two   string `toml:"two"`
	Few   string `toml:"few"`
	Many  string `toml:"many"`
	Other string `toml:"other"`
}

type fileCatalog struct {
	Language string             `toml:"language"`
	Messages map[string]string  `toml:"messages"`
	Plurals  map[string]message `toml:"plurals"`
}

var (
	catalogMu       sync.RWMutex
	defaultLanguage = LangZhCN
	catalog         = map[string]map[MessageKey]message{}
	templatePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)
)

func init() {
	resetCatalog()
}

func LoadDir(dir string) error {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	next := cloneCatalog()
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".toml" {
			continue
		}
		if err := loadFileInto(next, filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}

	catalogMu.Lock()
	catalog = next
	catalogMu.Unlock()
	return nil
}

func SetDefaultLanguage(lang string) {
	lang = canonicalLanguage(lang)
	if lang == "" {
		return
	}

	catalogMu.Lock()
	defaultLanguage = lang
	catalogMu.Unlock()
}

func Message(ctx fiber.Ctx, key MessageKey) string {
	return Render(LanguageFromCtx(ctx), key, nil)
}

func Messagef(ctx fiber.Ctx, key MessageKey, args ...any) string {
	return fmt.Sprintf(Message(ctx, key), args...)
}

func T(ctx fiber.Ctx, key MessageKey, data Data) string {
	return Render(LanguageFromCtx(ctx), key, data)
}

func Plural(ctx fiber.Ctx, key MessageKey, count int, data Data) string {
	return PluralByLang(LanguageFromCtx(ctx), key, count, data)
}

func MessageByLang(lang string, key MessageKey) string {
	return Render(lang, key, nil)
}

func Render(lang string, key MessageKey, data Data) string {
	msg, ok := findMessage(lang, key)
	if !ok {
		return string(key)
	}
	return renderTemplate(msg.Other, data)
}

func PluralByLang(lang string, key MessageKey, count int, data Data) string {
	msg, ok := findMessage(lang, key)
	if !ok {
		return string(key)
	}

	if data == nil {
		data = Data{}
	}
	data["count"] = count
	return renderTemplate(selectPlural(msg, count), data)
}

func LanguageFromCtx(ctx fiber.Ctx) string {
	return MatchLanguage(ctx.Get("Accept-Language"))
}

func MatchLanguage(header string) string {
	catalogMu.RLock()
	defer catalogMu.RUnlock()

	for _, candidate := range parseAcceptLanguage(header) {
		if _, ok := catalog[candidate]; ok {
			return candidate
		}
		base := strings.Split(candidate, "-")[0]
		for lang := range catalog {
			if strings.EqualFold(strings.Split(lang, "-")[0], base) {
				return lang
			}
		}
	}
	return defaultLanguage
}

func loadFileInto(target map[string]map[MessageKey]message, filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	file := fileCatalog{}
	if err := toml.Unmarshal(content, &file); err != nil {
		return fmt.Errorf("load i18n file %s: %w", filename, err)
	}

	lang := canonicalLanguage(file.Language)
	if lang == "" {
		lang = canonicalLanguage(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)))
	}
	if lang == "" {
		return fmt.Errorf("load i18n file %s: missing language", filename)
	}

	if _, ok := target[lang]; !ok {
		target[lang] = map[MessageKey]message{}
	}
	for key, value := range file.Messages {
		target[lang][MessageKey(key)] = message{Other: value}
	}
	for key, value := range file.Plurals {
		target[lang][MessageKey(key)] = value
	}
	return nil
}

func findMessage(lang string, key MessageKey) (message, bool) {
	catalogMu.RLock()
	defer catalogMu.RUnlock()

	for _, candidate := range []string{MatchLanguageLocked(lang), defaultLanguage, LangZhCN} {
		if messages, ok := catalog[candidate]; ok {
			if msg, ok := messages[key]; ok {
				return msg, true
			}
		}
	}
	return message{}, false
}

func MatchLanguageLocked(header string) string {
	for _, candidate := range parseAcceptLanguage(header) {
		if _, ok := catalog[candidate]; ok {
			return candidate
		}
		base := strings.Split(candidate, "-")[0]
		for lang := range catalog {
			if strings.EqualFold(strings.Split(lang, "-")[0], base) {
				return lang
			}
		}
	}
	return defaultLanguage
}

func parseAcceptLanguage(header string) []string {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil
	}

	parts := strings.Split(header, ",")
	languages := make([]string, 0, len(parts))
	for _, part := range parts {
		lang := strings.TrimSpace(strings.Split(part, ";")[0])
		lang = canonicalLanguage(lang)
		if lang != "" && lang != "*" {
			languages = append(languages, lang)
		}
	}
	return languages
}

func canonicalLanguage(lang string) string {
	lang = strings.TrimSpace(strings.ReplaceAll(lang, "_", "-"))
	if lang == "" {
		return ""
	}

	parts := strings.Split(lang, "-")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToLower(part)
			continue
		}
		if len(part) == 2 {
			parts[i] = strings.ToUpper(part)
			continue
		}
		parts[i] = strings.ToLower(part)
	}
	return strings.Join(parts, "-")
}

func selectPlural(msg message, count int) string {
	switch {
	case count == 0 && msg.Zero != "":
		return msg.Zero
	case count == 1 && msg.One != "":
		return msg.One
	case count == 2 && msg.Two != "":
		return msg.Two
	case count > 1 && count < 5 && msg.Few != "":
		return msg.Few
	case count >= 5 && msg.Many != "":
		return msg.Many
	case msg.Other != "":
		return msg.Other
	case msg.One != "":
		return msg.One
	case msg.Zero != "":
		return msg.Zero
	default:
		return ""
	}
}

func renderTemplate(tmpl string, data Data) string {
	if data == nil {
		return tmpl
	}
	return templatePattern.ReplaceAllStringFunc(tmpl, func(match string) string {
		name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}"))
		value, ok := data[name]
		if !ok {
			return match
		}
		return valueToString(value)
	})
}

func valueToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return fmt.Sprint(v)
	}
}

func cloneCatalog() map[string]map[MessageKey]message {
	catalogMu.RLock()
	defer catalogMu.RUnlock()

	next := map[string]map[MessageKey]message{}
	for lang, messages := range catalog {
		next[lang] = map[MessageKey]message{}
		for key, msg := range messages {
			next[lang][key] = msg
		}
	}
	return next
}

func resetCatalog() {
	catalogMu.Lock()
	defer catalogMu.Unlock()

	defaultLanguage = LangZhCN
	catalog = map[string]map[MessageKey]message{
		LangZhCN: {
			ErrBadRequest:            {Other: "请求错误，请检查参数"},
			ErrRequestTimeout:        {Other: "请求超时，请重新尝试"},
			ErrServerInternal:        {Other: "服务器内部错误，请稍后重试"},
			ErrServerContactAdmin:    {Other: "服务器异常，请联系管理员"},
			ErrSMSProvider:           {Other: "短信服务暂时不可用，请稍后重试"},
			ErrSMSCodeInvalid:        {Other: "验证码错误或不存在，请重试"},
			ErrPhoneHasNoAccount:     {Other: "该手机号还没有账号，请使用短信登录自动注册首个账号"},
			ErrNoPendingLoginAccount: {Other: "没有待选择的登录账号，请重新验证码登录"},
			ErrAccountNotBelongPhone: {Other: "账号不属于本次验证的手机号"},
			ErrUnauthorized:          {Other: "请先登录"},
			ErrUserNotFound:          {Other: "用户不存在"},
			ErrUserInfoConflict:      {Other: "用户信息已存在冲突，请更换后重试"},
			ErrAccountBanned:         {Other: "账号被封禁，无法继续操作"},
			ErrAccountClosed:         {Other: "账号已注销，无法继续操作"},
			ErrAccountDisabled:       {Other: "账号被临时禁用，无法继续操作"},
			ErrAccountUnavailable:    {Other: "账号状态异常，无法继续操作"},
			ErrPhoneNotBound:         {Other: "先绑定手机号再操作哦～"},
			ErrPasswordNotSet:        {Other: "先设置密码再操作哦～"},
			ErrCertificationRequired: {Other: "先完成账号认证再操作哦～"},
			ErrInvalidReferrer:       {Other: "操作失败啦，请再次打开页面重新操作"},
			ErrFileNotShared:         {Other: "此文件并未公开分享，请检查登录账号"},
			ErrFileBlockedByOwner:    {Other: "由于对方权限设置，无法查看此文件"},
			ErrBlockedByRelation:     {Other: "由于对方权限设置，无法查看此内容"},
			ErrPunished:              {Other: "账号封禁期间无法执行此操作"},
			ErrTargetPunished:        {Other: "该用户存在违规记录，部分功能受限"},
			ErrContentBlocked:        {Other: "内容因违规被屏蔽，无法显示"},
			ErrNeedLogin:             {Other: "请先登录后再查看"},

			UserClosedUsername: {Other: "已注销的账号"},
			UserClosedBio:      {Other: "用户已注销。虽然我不知道 TA 在三次元过得好不好，但我替这个账号感谢你最后的回眸。 —— 站长留"},

			OKSMSSent:               {Other: "发送成功"},
			OKLogin:                 {Other: "登录成功"},
			OKRegisterAndLogin:      {Other: "注册并登录成功"},
			OKNewAccountAndLogin:    {Other: "新账号注册并登录成功"},
			OKChooseLoginAccount:    {Other: "请选择要登录的账号"},
			OKAccountList:           {Other: "获取成功"},
			OKLogout:                {Other: "退出登录成功"},
			OKUserInfoUpdated:       {Other: "用户信息修改成功"},
			OKAvatarUpdated:         {Other: "头像设置成功"},
			OKAvatarFetched:         {Other: "获取成功"},
			OKMomentPublished:       {Other: "发布成功"},
			OKMomentUpdated:         {Other: "修改成功"},
			OKMomentDetailFetched:   {Other: "获取成功"},
			OKBlockUpdated:          {Other: "操作成功"},
			OKContentList:           {Other: "获取成功"},
			OKUserInfoFetched:       {Other: "获取成功"},
			OKCommentCreated:        {Other: "评论成功"},
			OKReplyCreated:          {Other: "回复成功"},
			OKConversationFetched:   {Other: "获取成功"},
			OKPunishmentsFetched:    {Other: "获取成功"},
			OKRelationList:          {Other: "获取成功"},
			OKInteract:              {Other: "操作成功"},
			OKArticleFetched:        {Other: "获取成功"},
			OKArticleFetchedPartial: {Other: "登录后查看完整内容"},
			OKArticleNeedSegments:   {Other: "文章过长，请使用分段接口获取正文"},
			OKArticleSegmentFetched: {Other: "获取成功"},
			OKFeedFetched:           {Other: "获取成功"},
			SMSActionLoginRegister:  {Other: "登录或注册"},
		},
	}
}
