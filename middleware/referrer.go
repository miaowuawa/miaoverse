package middleware

import (
	"net"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"miaoverse/consts"
	"miaoverse/model/dto/resp"
	"miaoverse/service/i18n"
)

type ReferrerConfig struct {
	AllowedHosts     []string
	AllowedOrigins   []string
	RejectSameHost   bool
	RejectNoReferrer bool
	RejectMobileUA   bool
	Status           int
	Message          i18n.MessageKey
}

func RequireReferrer(config ...ReferrerConfig) fiber.Handler {
	cfg := defaultReferrerConfig()
	if len(config) > 0 {
		cfg = mergeReferrerConfig(cfg, config[0])
	}

	allowedHosts := normalizeHostSet(cfg.AllowedHosts)
	allowedOrigins := normalizeOriginSet(cfg.AllowedOrigins)

	return func(ctx fiber.Ctx) error {
		if !cfg.RejectMobileUA && consts.RegexpUaWap.MatchString(ctx.UserAgent()) {
			return ctx.Next()
		}

		referrer := strings.TrimSpace(ctx.Get("Referer"))
		if referrer == "" {
			if !cfg.RejectNoReferrer {
				return ctx.Next()
			}
			return rejectReferrer(ctx, cfg)
		}

		referrerURL, err := url.Parse(referrer)
		if err != nil || referrerURL.Scheme == "" || referrerURL.Host == "" {
			return rejectReferrer(ctx, cfg)
		}

		if allowedOrigins[normalizeOrigin(referrerURL)] {
			return ctx.Next()
		}

		referrerHost := normalizeHost(referrerURL.Host)
		if allowedHosts[referrerHost] {
			return ctx.Next()
		}

		if !cfg.RejectSameHost && referrerHost == normalizeHost(ctx.Host()) {
			return ctx.Next()
		}

		return rejectReferrer(ctx, cfg)
	}
}

func defaultReferrerConfig() ReferrerConfig {
	return ReferrerConfig{
		Status:  fiber.StatusForbidden,
		Message: i18n.ErrInvalidReferrer,
	}
}

func mergeReferrerConfig(base, override ReferrerConfig) ReferrerConfig {
	base.AllowedHosts = override.AllowedHosts
	base.AllowedOrigins = override.AllowedOrigins
	base.RejectSameHost = override.RejectSameHost
	base.RejectNoReferrer = override.RejectNoReferrer
	base.RejectMobileUA = override.RejectMobileUA
	if override.Status != 0 {
		base.Status = override.Status
	}
	if override.Message != "" {
		base.Message = override.Message
	}
	return base
}

func normalizeHostSet(hosts []string) map[string]bool {
	result := make(map[string]bool, len(hosts))
	for _, host := range hosts {
		host = normalizeHost(host)
		if host != "" {
			result[host] = true
		}
	}
	return result
}

func normalizeOriginSet(origins []string) map[string]bool {
	result := make(map[string]bool, len(origins))
	for _, origin := range origins {
		parsed, err := url.Parse(strings.TrimSpace(origin))
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			continue
		}
		result[normalizeOrigin(parsed)] = true
	}
	return result
}

func normalizeOrigin(value *url.URL) string {
	return strings.ToLower(value.Scheme) + "://" + normalizeHost(value.Host)
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return ""
	}

	if strings.Contains(host, "://") {
		parsed, err := url.Parse(host)
		if err == nil {
			host = parsed.Host
		}
	}

	name, port, err := net.SplitHostPort(host)
	if err == nil {
		return strings.Trim(strings.ToLower(name), "[]") + ":" + port
	}
	return strings.Trim(host, "[]")
}

func rejectReferrer(ctx fiber.Ctx, cfg ReferrerConfig) error {
	return ctx.Status(cfg.Status).JSON(resp.CodeWithMsg{
		Code: cfg.Status,
		Msg:  i18n.Message(ctx, cfg.Message),
	})
}
