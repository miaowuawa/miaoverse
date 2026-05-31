package validate

import "miaoverse/consts"

const (
	UATypeUnknown string = "unknown" // 未知类型
	UATypePC             = "pc"      // PC端
	UATypeWAP            = "wap"     // 移动端/WAP
	UATypeBot            = "bot"     // 爬虫/Bot
)

// ParseUA 解析UA字符串，返回对应的类型
func ParseUA(ua string) string {
	// 第一步：判断是否为爬虫/Bot（优先级最高）
	if consts.RegexpUaBot.MatchString(ua) {
		return UATypeBot
	}

	// 第二步：判断是否为移动端/WAP
	if consts.RegexpUaBot.MatchString(ua) {
		return UATypeWAP
	}

	// 第三步：默认归为PC端（无法识别的UA按PC处理）
	return UATypePC
}
