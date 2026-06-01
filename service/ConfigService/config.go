package ConfigService

import (
	"fmt"
	"miaoverse/model/server/conf"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

//TODO:后面换成NACOS

// InitConfig 初始化配置（本地文件模式）—— 后续迁移Nacos只改这个函数
func InitConfig(configPath string) (error, *conf.AppConfig) {
	// 配置viper加载本地yaml文件
	GlobalConfig := &conf.AppConfig{}
	viper.SetConfigName("config")   // 配置文件名（无后缀）
	viper.SetConfigType("yaml")     // 配置格式
	viper.AddConfigPath(configPath) // 配置文件所在目录（可加多个）

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取本地配置文件失败: %v", err), &conf.AppConfig{}
	}

	// 解析到全局配置结构体
	if err := viper.Unmarshal(&GlobalConfig, func(config *mapstructure.DecoderConfig) {
		config.TagName = "yaml"
	}); err != nil {
		return fmt.Errorf("解析配置失败: %v", err), &conf.AppConfig{}
	}

	// （可选）监听本地配置文件变化，实现简单热更新
	//viper.WatchConfig()
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	fmt.Println("本地配置文件更新，重新加载...")
	//	if err := viper.Unmarshal(&GlobalConfig); err != nil {
	//		fmt.Printf("配置热更新失败: %v\n", err)
	//	}
	//})

	fmt.Println("本地配置初始化成功！")
	return nil, GlobalConfig
}
