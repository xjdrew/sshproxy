package webhook

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config webhook 服务配置
type Config struct {
	Listen string       `yaml:"listen"`
	Users  []UserConfig `yaml:"users"`
}

// UserConfig 用户配置
type UserConfig struct {
	Username  string            `yaml:"username"`
	Password  string            `yaml:"password"`
	PublicKey string            `yaml:"publicKey,omitempty"`
	Metadata  map[string]string `yaml:"metadata"`
}

// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	if config.Listen == "" {
		config.Listen = ":8080"
	}

	return &config, nil
}

// GetUser 根据用户名获取用户配置
func (c *Config) GetUser(username string) *UserConfig {
	for i := range c.Users {
		if c.Users[i].Username == username {
			return &c.Users[i]
		}
	}
	return nil
}
