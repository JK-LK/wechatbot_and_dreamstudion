package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

// Configuration 项目配置
type Configuration struct {
	// gtp apikey
	ApiKey string `json:"api_key"`
	// GPT模型
	Model string `json:"model"`
	// 自动通过好友
	AutoPass bool `json:"auto_pass"`
	// 会话超时时间
	SessionTimeout time.Duration `json:"session_timeout"`
	// 清空会话口令
	SessionClearToken string `json:"session_clear_token"`
	// dreamstdio apikey
	DreamStdioApiKey string `json:"dreamstdio_api_key"`
	// dreamstdio模型名称
	EngineId string `json:"engine_id"`
	// 图像生成的高度
	PicWidth uint `json:"picture_width"`
	// 图像生成的高度
	PicHeight uint `json:"picture_height"`
	// 图像生成迭代次数
	Steps uint `json:"steps"`
	// 图像生成系数
	CfgScale uint `json:"cfg_scale"`
	// 图像生成识别指令
	PictureToken string `json:"picture_token"`
}

var config *Configuration
var once sync.Once

// LoadConfig 加载配置
func LoadConfig() *Configuration {
	once.Do(func() {
		// 从文件中读取
		config = &Configuration{}
		f, err := os.Open("config.json")
		if err != nil {
			log.Fatalf("open config err: %v", err)
			return
		}
		defer f.Close()
		encoder := json.NewDecoder(f)
		err = encoder.Decode(config)
		if err != nil {
			log.Fatalf("decode config err: %v", err)
			return
		}

		// 如果环境变量有配置，读取环境变量
		ApiKey := os.Getenv("ApiKey")
		AutoPass := os.Getenv("AutoPass")
		if ApiKey != "" {
			config.ApiKey = ApiKey
		}
		if AutoPass == "true" {
			config.AutoPass = true
		}
	})
	return config
}
