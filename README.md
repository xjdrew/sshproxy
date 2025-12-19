# SSH Proxy for Kubernetes

基于 [ContainerSSH](https://containerssh.io/) 的 SSH 代理服务，允许用户通过 SSH 直接连接到 Kubernetes 集群中的 Pod 容器。

## ✨ 特性

- 🔐 **灵活的认证方式**：支持密码认证和 SSH 公钥认证
- 🎯 **精确的容器映射**：每个用户可以映射到特定的 Kubernetes Pod 和容器
- 🚀 **零侵入性**：使用 ContainerSSH 的 persistent 模式，连接到已存在的 Pod，无需创建新容器
- 📝 **详细的日志**：完整的认证和连接日志，便于调试和审计
- 🔧 **易于配置**：简单的 YAML 配置文件，支持热重载
- 🧪 **完整的测试**：包含单元测试和集成测试

## 📋 前置要求

- Go 1.21 或更高版本
- Kubernetes 集群访问权限
- kubectl 配置正确（用于本地开发测试）

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd sshproxy
```

### 2. 生成 SSH Host Key

首次使用需要生成 SSH host key：

```bash
ssh-keygen -t rsa -b 2048 -f ssh_host_rsa_key -N "" -C "containerssh@sshproxy"
chmod 600 ssh_host_rsa_key
```

### 3. 配置 Kubernetes 连接

编辑 `config.yaml`，配置 Kubernetes API Server 连接信息：

```yaml
kubernetes:
  connection:
    host: "https://your-k8s-api-server:6443"
    cacertFile: "/path/to/ca.crt"
    certFile: "/path/to/client.crt"
    keyFile: "/path/to/client.key"
```

**提示**：可以从 `~/.kube/config` 中提取这些信息。

### 4. 配置用户和 Pod 映射

编辑 `webhook.yaml`，添加用户和对应的容器映射：

```yaml
listen: ":8080"

users:
  - username: "developer"
    password: "dev_password"
    metadata:
      KUBERNETES_POD_NAMESPACE: "default"
      KUBERNETES_POD_NAME: "my-app-pod"
      KUBERNETES_CONTAINER_NAME: "app"
```

### 5. 构建项目

```bash
make build
```

或手动构建：

```bash
go build -o bin/containerssh ./cmd/containerssh
go build -o bin/sshhook ./cmd/sshhook
```

### 6. 启动服务

**方式 1：使用 Makefile（推荐）**

```bash
# 启动所有服务
make run

# 或分别启动
make run-hook    # 启动 Webhook 服务
make run-ssh     # 启动 SSH 服务
```

**方式 2：手动启动**

在两个终端中分别运行：

```bash
# 终端 1：启动 Webhook 服务
./bin/sshhook --config webhook.yaml

# 终端 2：启动 ContainerSSH
./bin/containerssh --config config.yaml
```

### 7. 连接测试

使用 SSH 客户端连接：

```bash
# 密码认证
ssh developer@localhost -p 2222

# 公钥认证
ssh -i ~/.ssh/your_private_key developer@localhost -p 2222
```

## 📁 项目结构

```
sshproxy/
├── cmd/
│   ├── containerssh/          # ContainerSSH 主程序入口
│   └── sshhook/               # Webhook 服务入口
├── pkg/
│   └── webhook/               # Webhook 实现
│       ├── config.go          # 配置加载
│       ├── server.go          # HTTP 服务器和认证逻辑
│       ├── config_test.go     # 配置测试
│       └── server_test.go     # 服务器测试
├── config.yaml                # ContainerSSH 配置文件
├── webhook.yaml               # Webhook 服务配置文件
├── .gitignore                 # Git 忽略文件
├── Makefile                   # 构建脚本
├── go.mod                     # Go 模块定义
├── go.sum                     # Go 依赖锁定
└── README.md                  # 项目文档
```

## ⚙️ 配置详解

### ContainerSSH 配置 (config.yaml)

详细的配置说明请参考 `config.yaml` 文件中的注释。主要配置项：

- **ssh**: SSH 服务配置（监听地址、host key、banner 等）
- **auth**: 认证配置（webhook URL、认证方式等）
- **configserver**: 配置服务器（用于动态配置后端）
- **backend**: 后端类型（kubernetes）
- **kubernetes**: Kubernetes 连接和 Pod 配置
- **log**: 日志配置

### Webhook 配置 (webhook.yaml)

详细的配置说明请参考 `webhook.yaml` 文件中的注释。主要配置项：

- **listen**: Webhook 服务监听地址
- **users**: 用户列表
  - **username**: SSH 用户名
  - **password**: 密码（可选）
  - **publicKey**: SSH 公钥（可选）
  - **metadata**: Pod 映射信息
    - **KUBERNETES_POD_NAMESPACE**: Pod 所在的 namespace
    - **KUBERNETES_POD_NAME**: Pod 名称
    - **KUBERNETES_CONTAINER_NAME**: 容器名称（可选）

## 🔐 认证方式

### 密码认证

在 `webhook.yaml` 中配置用户密码：

```yaml
users:
  - username: "user1"
    password: "secure_password"
    metadata:
      KUBERNETES_POD_NAMESPACE: "default"
      KUBERNETES_POD_NAME: "my-pod"
```

连接：

```bash
ssh user1@your-server -p 2222
# 输入密码：secure_password
```

### 公钥认证（推荐）

1. 生成 SSH 密钥对：

```bash
ssh-keygen -t rsa -b 2048 -f ~/.ssh/sshproxy_key
```

2. 在 `webhook.yaml` 中配置公钥：

```yaml
users:
  - username: "user1"
    publicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ... user@host"
    metadata:
      KUBERNETES_POD_NAMESPACE: "default"
      KUBERNETES_POD_NAME: "my-pod"
```

3. 连接：

```bash
ssh -i ~/.ssh/sshproxy_key user1@your-server -p 2222
```

## 🧪 开发和测试

### 运行测试

```bash
# 运行所有测试
make test

# 或手动运行
go test ./...

# 运行特定包的测试
go test -v ./pkg/webhook/...

# 运行测试并显示覆盖率
go test -cover ./...
```

### 代码检查

```bash
# 代码格式化
go fmt ./...

# 代码检查
go vet ./...

# 使用 golangci-lint（如果已安装）
golangci-lint run
```

### 本地开发配置

项目支持本地开发配置文件（不会提交到 Git）：

- `.local-ssh.yaml`: 本地 ContainerSSH 配置
- `.local-hook.yaml`: 本地 Webhook 配置

使用本地配置启动：

```bash
./bin/sshhook --config .local-hook.yaml
./bin/containerssh --config .local-ssh.yaml
```

## 📊 日志和调试

### 查看日志

ContainerSSH 和 Webhook 服务都会输出详细的日志：

```bash
# ContainerSSH 日志
[SSH] Connection from 127.0.0.1:xxxxx
[Auth] Password authentication request for user: developer
[Auth] ✓ Password authentication successful for user: developer
[Config] Configuration returned - namespace=default, pod=my-app-pod

# Webhook 日志
[Password] Request received - username=developer
[Password] ✓ Authentication successful - username=developer
[Config] Request received - username=developer
[Config] ✓ Configuration returned - namespace=default, pod=my-app-pod
```

### 调试模式

修改 `config.yaml` 中的日志级别：

```yaml
log:
  level: debug  # 可选：debug, info, warning, error
```

## 🔧 常见问题

### 1. 连接被拒绝

**问题**：`ssh: connect to host localhost port 2222: Connection refused`

**解决**：
- 确保 ContainerSSH 服务正在运行
- 检查端口是否被占用：`lsof -i :2222`
- 检查防火墙设置

### 2. 认证失败

**问题**：`Permission denied (publickey,password)`

**解决**：
- 检查 `webhook.yaml` 中的用户配置
- 确保 Webhook 服务正在运行
- 查看 Webhook 日志确认认证请求

### 3. 无法连接到 Pod

**问题**：`Failed to connect to pod`

**解决**：
- 确保 Pod 存在且正在运行：`kubectl get pods -n <namespace>`
- 检查 Kubernetes 连接配置
- 确保有足够的 RBAC 权限
- 验证 Pod 名称和 namespace 配置正确

### 4. SSH Host Key 警告

**问题**：`WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!`

**解决**：
- 删除旧的 host key：`ssh-keygen -R [localhost]:2222`
- 或编辑 `~/.ssh/known_hosts` 删除对应行

## 🚀 生产部署建议

1. **使用公钥认证**：禁用密码认证，只使用 SSH 公钥
2. **配置 TLS**：为 Webhook 服务配置 HTTPS
3. **限制访问**：使用防火墙限制 SSH 端口访问
4. **日志审计**：配置日志收集和审计
5. **监控告警**：配置 Prometheus 监控和告警
6. **定期更新**：保持 ContainerSSH 和依赖库更新
7. **RBAC 权限**：为 ContainerSSH 配置最小权限的 Kubernetes RBAC

## 📚 参考文档

- [ContainerSSH 官方文档](https://containerssh.io/)
- [ContainerSSH GitHub](https://github.com/ContainerSSH/ContainerSSH)
- [Kubernetes API 文档](https://kubernetes.io/docs/reference/)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT License - 详见 LICENSE 文件