package conf

type AppConfig struct {
	Server struct {
		Port     int    `yaml:"port"`
		LogLevel string `yaml:"log_level"`
	} `yaml:"server"`
	Sql struct {
		Host            string `yaml:"host"`
		Port            int    `yaml:"port"`
		Username        string `yaml:"username"`
		Password        string `yaml:"password"`
		DbName          string `yaml:"db_name"`
		MaxOpenConns    int    `yaml:"max_open_conns"`
		MaxIdleConns    int    `yaml:"max_idle_conns"`
		ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
		ConnMaxIdletime int    `yaml:"conn_max_idle_time"`
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
	Session struct {
		DB int `yaml:"dbnum"`
	} `yaml:"session"`
	SMS struct {
		RateLimit int `yaml:"rate_limit"` // 短信发送速率限制，单位：条/小时
	} `yaml:"smsreq"`
	I18n struct {
		Dir             string `yaml:"dir"`
		DefaultLanguage string `yaml:"default_language"`
	} `yaml:"i18n"`
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
