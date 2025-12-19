package webhook

import (
	"os"
	"testing"
)

// TestLoadConfig_Success 测试成功加载配置
func TestLoadConfig_Success(t *testing.T) {
	// 创建临时配置文件
	content := `listen: ":9090"
users:
  - username: "user1"
    password: "pass1"
    metadata:
      namespace: "default"
      pod: "pod1"
      container: "container1"
  - username: "user2"
    password: "pass2"
    publicKey: "ssh-rsa AAAAB3... user2@example.com"
    metadata:
      namespace: "test"
      pod: "pod2"
      container: "container2"
`
	tmpfile, err := os.CreateTemp("", "webhook-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	// 加载配置
	config, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置
	if config.Listen != ":9090" {
		t.Errorf("Expected listen ':9090', got '%s'", config.Listen)
	}

	if len(config.Users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(config.Users))
	}

	// 验证第一个用户
	user1 := config.Users[0]
	if user1.Username != "user1" {
		t.Errorf("Expected username 'user1', got '%s'", user1.Username)
	}
	if user1.Password != "pass1" {
		t.Errorf("Expected password 'pass1', got '%s'", user1.Password)
	}
	if user1.Metadata["namespace"] != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", user1.Metadata["namespace"])
	}

	// 验证第二个用户
	user2 := config.Users[1]
	if user2.Username != "user2" {
		t.Errorf("Expected username 'user2', got '%s'", user2.Username)
	}
	if user2.PublicKey == "" {
		t.Error("Expected public key to be set for user2")
	}
}

// TestLoadConfig_DefaultListen 测试默认监听地址
func TestLoadConfig_DefaultListen(t *testing.T) {
	content := `users:
  - username: "user1"
    password: "pass1"
    metadata:
      namespace: "default"
`
	tmpfile, err := os.CreateTemp("", "webhook-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	config, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证默认监听地址
	if config.Listen != ":8080" {
		t.Errorf("Expected default listen ':8080', got '%s'", config.Listen)
	}
}

// TestLoadConfig_FileNotFound 测试文件不存在
func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestLoadConfig_InvalidYAML 测试无效的 YAML
func TestLoadConfig_InvalidYAML(t *testing.T) {
	content := `invalid: yaml: content:
  - this is not valid
    yaml syntax
`
	tmpfile, err := os.CreateTemp("", "webhook-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpfile.Close()

	_, err = LoadConfig(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// TestGetUser_Found 测试查找存在的用户
func TestGetUser_Found(t *testing.T) {
	config := &Config{
		Users: []UserConfig{
			{Username: "user1", Password: "pass1"},
			{Username: "user2", Password: "pass2"},
			{Username: "user3", Password: "pass3"},
		},
	}

	user := config.GetUser("user2")
	if user == nil {
		t.Fatal("Expected to find user2, got nil")
	}

	if user.Username != "user2" {
		t.Errorf("Expected username 'user2', got '%s'", user.Username)
	}
	if user.Password != "pass2" {
		t.Errorf("Expected password 'pass2', got '%s'", user.Password)
	}
}

// TestGetUser_NotFound 测试查找不存在的用户
func TestGetUser_NotFound(t *testing.T) {
	config := &Config{
		Users: []UserConfig{
			{Username: "user1", Password: "pass1"},
			{Username: "user2", Password: "pass2"},
		},
	}

	user := config.GetUser("nonexistent")
	if user != nil {
		t.Errorf("Expected nil for non-existent user, got %+v", user)
	}
}

// TestGetUser_EmptyList 测试空用户列表
func TestGetUser_EmptyList(t *testing.T) {
	config := &Config{
		Users: []UserConfig{},
	}

	user := config.GetUser("anyuser")
	if user != nil {
		t.Errorf("Expected nil for empty user list, got %+v", user)
	}
}
