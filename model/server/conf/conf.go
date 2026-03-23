package conf

type AppConfig struct {
	Server struct {
		Port     int    `yaml:"port"`
		LogLevel string `yaml:"log_level"`
	} `yaml:"server"`
	Sql struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"bcrypthash"`
		DbName   string `yaml:"db_name"`
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"bcrypthash"`
	} `yaml:"database"`
	SMS struct {
		RateLimit int `yaml:"rate_limit"` // 短信发送速率限制，单位：条/小时
	} `yaml:"smsreq"`
	SmsBao struct {
		Gateway  string `yaml:"gateway"`
		Username string `yaml:"username"`
		Passwd   string `yaml:"passwd"`
		Head     string `yaml:"head"`
		DB       int    `yaml:"dbnum"`
	} `yaml:"smsbao"`
	Cache struct {
		DB int `yaml:"dbnum"`
	}
}
