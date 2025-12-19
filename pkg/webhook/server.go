package webhook

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.containerssh.io/containerssh/auth"
	"go.containerssh.io/containerssh/config"
	"go.containerssh.io/containerssh/metadata"
	"golang.org/x/crypto/ssh"
	v1 "k8s.io/api/core/v1"
)

// Server webhook HTTP 服务器
type Server struct {
	config     *Config
	httpServer *http.Server
}

// AuthResponse 认证响应（使用 ContainerSSH 的 ResponseBody）
// 注意：auth.ResponseBody 已经包含了 Success 和 AuthenticatedUsername 字段
// 以及 metadata.ConnectionAuthenticatedMetadata（包含 Metadata、Environment、Files）
type AuthResponse = auth.ResponseBody

// NewServer 创建新的 webhook 服务器
func NewServer(config *Config) (*Server, error) {
	server := &Server{
		config: config,
	}

	// 注册路由
	http.HandleFunc("/config", server.handleConfig)         // Config 接口
	http.HandleFunc("/password", server.handlePasswordAuth) // 密码认证
	http.HandleFunc("/pubkey", server.handlePublicKeyAuth)  // 公钥认证

	server.httpServer = &http.Server{
		Addr:         config.Listen,
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// handlePasswordAuth 处理密码认证
func (s *Server) handlePasswordAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[Password Auth] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.PasswordAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Password Auth] Failed to decode request: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("[Password Auth] Request received - username=%s, remoteAddress=%s, connectionId=%s",
		req.Username, req.RemoteAddress, req.ConnectionID)

	// 注意：根据 ContainerSSH auth 协议，虽然 Password 字段类型是 []byte，
	// 但 JSON 中的字段名是 passwordBase64，实际传输的是 Base64 编码的密码
	// 所以 req.Password 中存储的是 Base64 编码后的字符串，需要先转为字符串再解码
	passwordBase64 := string(req.Password)

	// 从 Base64 解码得到原始密码
	passwordBytes, err := base64.StdEncoding.DecodeString(passwordBase64)
	if err != nil {
		log.Printf("[Password Auth] Failed to decode password for user %s: %v", req.Username, err)
		s.sendAuthResponse(w, false, "", nil)
		return
	}
	password := string(passwordBytes)

	// 查找用户
	user := s.config.GetUser(req.Username)
	if user == nil {
		log.Printf("[Password Auth] User not found: %s", req.Username)
		s.sendAuthResponse(w, false, "", nil)
		return
	}

	// 验证密码
	if user.Password != password {
		log.Printf("[Password Auth] Invalid password for user: %s", req.Username)
		s.sendAuthResponse(w, false, "", nil)
		return
	}

	log.Printf("[Password Auth] ✓ Authentication successful - username=%s, namespace=%s, pod=%s, container=%s",
		req.Username, user.Metadata["namespace"], user.Metadata["pod"], user.Metadata["container"])
	s.sendAuthResponse(w, true, req.Username, user.Metadata)
}

// handlePublicKeyAuth 处理公钥认证
func (s *Server) handlePublicKeyAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[Public Key Auth] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.PublicKeyAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Public Key Auth] Failed to decode request: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("[Public Key Auth] Request received - username=%s, remoteAddress=%s, connectionId=%s",
		req.Username, req.RemoteAddress, req.ConnectionID)

	// 查找用户
	user := s.config.GetUser(req.Username)
	if user == nil {
		log.Printf("User not found: %s", req.Username)
		s.sendAuthResponse(w, false, "", nil)
		return
	}

	// 如果用户没有配置公钥，拒绝认证
	if user.PublicKey == "" {
		log.Printf("No public key configured for user: %s", req.Username)
		s.sendAuthResponse(w, false, "", nil)
		return
	}

	// 解析客户端公钥（SSH authorized key 格式，如 "ssh-rsa AAAAB3..."）
	clientPubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey.PublicKey))
	if err != nil {
		log.Printf("Failed to parse client public key: %v", err)
		s.sendAuthResponse(w, false, "", nil)
		return
	}

	// 解析配置中的公钥（支持 OpenSSH authorized_keys 格式）
	configPubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(user.PublicKey))
	if err != nil {
		log.Printf("Failed to parse config public key: %v", err)
		s.sendAuthResponse(w, false, "", nil)
		return
	}

	// 比较公钥（通过比较 Marshal 后的字节）
	if !bytes.Equal(clientPubKey.Marshal(), configPubKey.Marshal()) {
		log.Printf("Public key mismatch for user: %s", req.Username)
		s.sendAuthResponse(w, false, "", nil)
		return
	}

	log.Printf("Public key auth success: username=%s", req.Username)
	s.sendAuthResponse(w, true, req.Username, user.Metadata)
}

// 使用 ContainerSSH 官方的 config 类型
// config.Request 和 config.ResponseBody 已在 go.containerssh.io/containerssh/config 包中定义

// handleConfig 处理配置请求
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[Config] Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req config.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Config] Failed to decode request: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("[Config] Request received - username=%s, authenticatedUsername=%s, connectionId=%s",
		req.Username, req.AuthenticatedUsername, req.ConnectionID)

	// 查找用户
	user := s.config.GetUser(req.AuthenticatedUsername)
	if user == nil {
		log.Printf("[Config] User not found: %s", req.AuthenticatedUsername)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 获取集群配置
	clusterName := user.Metadata["KUBERNETES_CLUSTER"]
	if clusterName == "" {
		log.Printf("[Config] Missing cluster name for user: %s", req.AuthenticatedUsername)
		http.Error(w, "Missing cluster configuration", http.StatusBadRequest)
		return
	}

	cluster := s.config.GetCluster(clusterName)
	if cluster == nil {
		log.Printf("[Config] Cluster not found: %s", clusterName)
		http.Error(w, "Cluster not found", http.StatusNotFound)
		return
	}

	// 构建 Kubernetes 配置
	podName := user.Metadata["KUBERNETES_POD_NAME"]
	namespace := user.Metadata["KUBERNETES_POD_NAMESPACE"]
	containerName := user.Metadata["KUBERNETES_CONTAINER_NAME"]

	if podName == "" || namespace == "" {
		log.Printf("[Config] Missing pod configuration for user: %s", req.AuthenticatedUsername)
		http.Error(w, "Missing pod configuration", http.StatusBadRequest)
		return
	}

	// 构建 Kubernetes Pod 配置
	kubeConfig := config.KubernetesConfig{}

	// 设置集群连接信息
	kubeConfig.Connection.Host = cluster.Host
	kubeConfig.Connection.CAFile = cluster.CACertFile
	kubeConfig.Connection.CertFile = cluster.CertFile
	kubeConfig.Connection.KeyFile = cluster.KeyFile
	if cluster.BearerTokenFile != "" {
		kubeConfig.Connection.BearerTokenFile = cluster.BearerTokenFile
	}
	if cluster.ServerName != "" {
		kubeConfig.Connection.ServerName = cluster.ServerName
	}
	if cluster.QPS > 0 {
		kubeConfig.Connection.QPS = float32(cluster.QPS)
	}
	if cluster.Burst > 0 {
		kubeConfig.Connection.Burst = cluster.Burst
	}

	// 设置 Pod 配置
	kubeConfig.Pod.Metadata.Name = podName
	kubeConfig.Pod.Metadata.Namespace = namespace

	// 设置 shell 命令（默认使用 /bin/bash）
	kubeConfig.Pod.ShellCommand = []string{"/bin/bash"}

	// 在 persistent 模式下，禁用 ContainerSSH agent
	kubeConfig.Pod.DisableAgent = true

	// 如果指定了容器名，设置容器配置
	// 注意：这里只设置容器名称，ContainerSSH 会使用这个名称来 exec 到指定容器
	if containerName != "" {
		kubeConfig.Pod.Spec.Containers = []v1.Container{
			{Name: containerName},
		}
		// 设置要连接的容器索引为 0（第一个容器）
		kubeConfig.Pod.ConsoleContainerNumber = 0
	}

	// 构建完整的应用配置
	appConfig := config.AppConfig{
		Backend:    config.BackendKubernetes,
		Kubernetes: kubeConfig,
	}

	// 使用 ContainerSSH 官方的 ResponseBody 结构
	resp := config.ResponseBody{
		ConnectionAuthenticatedMetadata: req.ConnectionAuthenticatedMetadata,
		Config:                          appConfig,
	}

	log.Printf("[Config] ✓ Configuration returned - cluster=%s, namespace=%s, pod=%s, container=%s",
		clusterName, namespace, podName, containerName)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[Config] Failed to encode response: %v", err)
	}
}

// sendAuthResponse 发送认证响应
func (s *Server) sendAuthResponse(w http.ResponseWriter, success bool, username string, userMetadata map[string]string) {
	resp := auth.ResponseBody{
		Success: success,
	}

	// 如果认证成功，设置用户名和 metadata
	if success {
		resp.AuthenticatedUsername = username

		// 转换 metadata 格式为 ContainerSSH 要求的格式
		if userMetadata != nil {
			metadataMap := make(map[string]metadata.Value)
			for key, value := range userMetadata {
				metadataMap[key] = metadata.Value{
					Value:     value,
					Sensitive: false,
				}
			}
			resp.ConnectionAuthPendingMetadata.ConnectionMetadata.Metadata = metadataMap
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode auth response: %v", err)
	}
}
