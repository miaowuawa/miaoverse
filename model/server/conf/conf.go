package conf

type AppConfig struct {
	Server struct {
		Port     int    `yaml:"port"`
		LogLevel string `yaml:"log_level"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		DbName   string `yaml:"db_name"`
	} `yaml:"database"`
	SMS struct {
		Type      string `yaml:"type"`       //1-短信宝
		RateLimit int    `yaml:"rate_limit"` // 短信发送速率限制，单位：条/小时
	} `yaml:"sms"`
	SmsBao struct {
		Gateway  string `yaml:"gateway"`
		Username string `yaml:"username"`
		Passwd   string `yaml:"passwd"`
		Head     string `yaml:"head"`
	} `yaml:"smsbao"`
}
