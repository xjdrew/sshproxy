package webhook

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ClusterConfig Kubernetes 集群配置
type ClusterConfig struct {
	Name            string `yaml:"name"`            // 集群名称（唯一标识）
	Host            string `yaml:"host"`            // Kubernetes API Server 地址
	CACertFile      string `yaml:"cacertFile"`      // CA 证书文件路径
	CertFile        string `yaml:"certFile"`        // 客户端证书文件路径
	KeyFile         string `yaml:"keyFile"`         // 客户端密钥文件路径
	BearerTokenFile string `yaml:"bearerTokenFile"` // Bearer Token 文件路径（可选）
	ServerName      string `yaml:"serverName"`      // TLS 服务器名称（可选）
	QPS             int    `yaml:"qps"`             // QPS 限制（可选）
	Burst           int    `yaml:"burst"`           // Burst 限制（可选）
}

// Config webhook 服务配置
type Config struct {
	Listen   string          `yaml:"listen"`
	Clusters []ClusterConfig `yaml:"clusters"` // Kubernetes 集群配置列表
	Users    []UserConfig    `yaml:"users"`
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

// GetCluster 根据集群名称获取集群配置
func (c *Config) GetCluster(name string) *ClusterConfig {
	for i := range c.Clusters {
		if c.Clusters[i].Name == name {
			return &c.Clusters[i]
		}
	}
	return nil
}
