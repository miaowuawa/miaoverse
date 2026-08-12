package smsbao

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"miaoverse/model/apireq/sms/smsbaoreq"
	sconf "miaoverse/model/server/conf"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// smsBaoHTTPTimeout 短信宝网关请求超时，避免网关无响应时长期占用连接。
const smsBaoHTTPTimeout = 10 * time.Second

// smsBaoMaxResponseBytes 短信宝响应体读取上限，防止异常响应耗尽内存。
const smsBaoMaxResponseBytes = 4 << 10

// SendPhoneCaptcha 发送短信验证码
func (s *SmsBaoServant) SendPhoneCaptcha(phone string, captcha string, expire time.Duration, usage string) error {
	// 对密码进行 MD5 加密
	hasher := md5.New()
	_, writeString := io.WriteString(hasher, s.Password)
	if writeString != nil {
		return writeString
	}
	encryptedPassword := hex.EncodeToString(hasher.Sum(nil))

	// 构建短信内容
	content := fmt.Sprintf("%s您的验证码是：%s，有效期 %d 分钟。就算猫娘来你家也不要告诉她。用于%s，如果用途不符请勿输入。", s.Head, captcha, int(expire.Minutes()), usage)

	// 对内容进行 URL 编码
	encodedContent := url.QueryEscape(content)

	// 构建请求 URL
	requestURL := fmt.Sprintf("%s?u=%s&p=%s&m=%s&c=%s", s.Gateway, s.Username, encryptedPassword, phone, encodedContent)

	// 发送 HTTP 请求（带超时，避免网关无响应时长期占用连接）
	client := &http.Client{Timeout: smsBaoHTTPTimeout}
	resp, err := client.Get(requestURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	// 读取响应内容（限制大小，防止异常响应耗尽内存）
	body, err := io.ReadAll(io.LimitReader(resp.Body, smsBaoMaxResponseBytes))
	if err != nil {
		return err
	}

	// 解析响应
	result := &smsbaoreq.PhoneCaptchaRsp{}
	// 假设响应是简单的 text/plain 格式，按行分割
	lines := strings.Split(string(body), "\n")
	if len(lines) >= 1 {
		result.Code = lines[0]
		if len(lines) >= 2 {
			result.Msg = strings.Join(lines[1:], "\n")
		}
	}

	// 处理响应结果：详细错误只写服务端日志，不向调用方暴露短信宝内部状态
	switch result.Code {
	case "0":
		return nil
	case "30":
		log.Printf("[smsbao] 发送失败：密码错误")
		return errors.New("短信服务发送失败")
	case "40":
		log.Printf("[smsbao] 发送失败：账号不存在")
		return errors.New("短信服务发送失败")
	case "41":
		log.Printf("[smsbao] 发送失败：余额不足")
		return errors.New("短信服务发送失败")
	case "42":
		log.Printf("[smsbao] 发送失败：帐号过期")
		return errors.New("短信服务发送失败")
	case "43":
		log.Printf("[smsbao] 发送失败：IP地址限制")
		return errors.New("短信服务发送失败")
	case "50":
		log.Printf("[smsbao] 发送失败：内容含有敏感词")
		return errors.New("短信服务发送失败")
	case "51":
		log.Printf("[smsbao] 发送失败：手机号码不正确")
		return errors.New("短信服务发送失败")
	case "-1":
		log.Printf("[smsbao] 发送失败：手机号码不正确或缺少参数")
		return errors.New("短信服务发送失败")
	default:
		log.Printf("[smsbao] 发送失败：未知错误 code=%s msg=%s", result.Code, result.Msg)
		return errors.New("短信服务发送失败")
	}
}

// newSmsBaoClient 创建短信宝服务实例
func newSmsBaoClient(conf *sconf.AppConfig) *SmsBaoServant {
	return &SmsBaoServant{
		Gateway:  conf.SmsBao.Gateway,
		Username: conf.SmsBao.Username,
		Password: conf.SmsBao.Passwd,
		Head:     conf.SmsBao.Head,
	}
}

// SmsBaoServant 短信宝服务结构体
type SmsBaoServant struct {
	Gateway  string
	Username string
	Password string
	Head     string
}
