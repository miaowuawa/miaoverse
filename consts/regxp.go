package consts

import "regexp"

// 预编译正则表达式（只编译一次，提升性能）
var (
	// 匹配爬虫/Bot 关键词（覆盖常见爬虫标识）
	RegexpUaBot = regexp.MustCompile(`(?i)bot|crawl|spider|scanner|curl|wget|python-requests|java|node.js|httpclient|scrapy|phantomjs|headless`)

	// 匹配移动端/WAP 关键词（覆盖手机、平板、移动端浏览器）
	RegexpUaWap = regexp.MustCompile(`(?i)mobile|android|iphone|ipad|ios|webos|blackberry|iemobile|symbian|windows phone|ucbrowser|qqbrowser|miui|harmonyos`)
)
