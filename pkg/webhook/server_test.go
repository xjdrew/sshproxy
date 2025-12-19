package webhook

import (
"testing"
)

// 创建测试配置
func createTestConfig() *Config {
return &Config{
Listen: ":8080",
Users: []UserConfig{
{
Username:  "testuser",
Password:  "testpass",
PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDFzTZY7Jxj9Brdidms69Z8ZQzAy2WDU0mnd/oJIlK0hqUrooiPfE7jfOLBhPSuxq6s4xuUlGm1HwuAC7vQak3JfEVFktmddpQpjCvrQgAtMZzad+q29+4VmwgX2AAWozwRx2n4tZb9/XodQms/t3Gr4iTilzMLKQ/zfCrtg7R8watAEKuaIq5nt6yD5LzcVfst3wK8cchjhbr0LpnJgCe3t82O4o0tKUXjZgCHDEK1lH+AaruaG8zAo+AYpIntm4dA+oPv3ZAs5Tjip9lCp+FDT+XCoH7BmWZPapMaOKRFxYhwhu7fzZxTl6QimpP/nIOzTKvtbOIElLRHH3T7TY7p test@example.com",
Metadata: map[string]string{
"KUBERNETES_POD_NAMESPACE":  "default",
"KUBERNETES_POD_NAME":       "test-pod",
"KUBERNETES_CONTAINER_NAME": "test-container",
},
},
},
}
}

// TestNewServer 测试创建服务器
func TestNewServer(t *testing.T) {
config := createTestConfig()
server, err := NewServer(config)
if err != nil {
t.Fatalf("Failed to create server: %v", err)
}

if server == nil {
t.Fatal("Expected server to be created, got nil")
}

if server.config != config {
t.Error("Server config not set correctly")
}
}

// TestGetUser 测试获取用户
func TestGetUser(t *testing.T) {
config := createTestConfig()

// 测试找到用户
user := config.GetUser("testuser")
if user == nil {
t.Fatal("Expected to find testuser, got nil")
}

if user.Username != "testuser" {
t.Errorf("Expected username 'testuser', got '%s'", user.Username)
}

if user.Password != "testpass" {
t.Errorf("Expected password 'testpass', got '%s'", user.Password)
}

// 测试用户不存在
user = config.GetUser("nonexistent")
if user != nil {
t.Errorf("Expected nil for non-existent user, got %+v", user)
}
}

// TestUserMetadata 测试用户元数据
func TestUserMetadata(t *testing.T) {
config := createTestConfig()
user := config.GetUser("testuser")

if user == nil {
t.Fatal("Expected to find testuser")
}

// 验证元数据
if user.Metadata["KUBERNETES_POD_NAMESPACE"] != "default" {
t.Errorf("Expected namespace 'default', got '%s'", user.Metadata["KUBERNETES_POD_NAMESPACE"])
}

if user.Metadata["KUBERNETES_POD_NAME"] != "test-pod" {
t.Errorf("Expected pod 'test-pod', got '%s'", user.Metadata["KUBERNETES_POD_NAME"])
}

if user.Metadata["KUBERNETES_CONTAINER_NAME"] != "test-container" {
t.Errorf("Expected container 'test-container', got '%s'", user.Metadata["KUBERNETES_CONTAINER_NAME"])
}
}
