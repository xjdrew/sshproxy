# Kubernetes Backend 配置指南

## 概述

ContainerSSH 需要连接到 Kubernetes 集群才能 exec 到 pod。本文档说明如何配置 Kubernetes 认证信息。

## 获取 Kubernetes 认证信息

### 方法 1：使用本地 kubeconfig（推荐用于开发环境）

如果你的 `kubectl` 命令可以正常工作，可以从 kubeconfig 中提取认证信息：

```bash
# 1. 查看当前使用的 context
kubectl config current-context

# 2. 查看 kubeconfig 文件位置
echo $KUBECONFIG
# 或默认位置：~/.kube/config

# 3. 查看 API Server 地址
kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'

# 4. 提取证书和密钥
# CA 证书
kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d > /tmp/ca.crt

# 客户端证书
kubectl config view --raw --minify --flatten -o jsonpath='{.users[0].user.client-certificate-data}' | base64 -d > /tmp/client.crt

# 客户端密钥
kubectl config view --raw --minify --flatten -o jsonpath='{.users[0].user.client-key-data}' | base64 -d > /tmp/client.key
```

### 方法 2：使用 Bearer Token（推荐用于生产环境）

```bash
# 如果使用 token 认证
kubectl config view --raw --minify --flatten -o jsonpath='{.users[0].user.token}' > /tmp/token
```

## 配置 config.yaml

根据你的认证方式，更新 `config.yaml` 中的 `kubernetes.connection` 部分：

### 使用证书认证（x509）

```yaml
kubernetes:
  connection:
    host: "https://your-k8s-api-server:6443"  # 从步骤 3 获取
    cacertFile: "/tmp/ca.crt"                 # CA 证书路径
    certFile: "/tmp/client.crt"               # 客户端证书路径
    keyFile: "/tmp/client.key"                # 客户端密钥路径
```

### 使用 Bearer Token 认证

```yaml
kubernetes:
  connection:
    host: "https://your-k8s-api-server:6443"
    cacertFile: "/tmp/ca.crt"
    bearerTokenFile: "/tmp/token"
```

### 使用内嵌证书（不推荐，仅用于测试）

```yaml
kubernetes:
  connection:
    host: "https://your-k8s-api-server:6443"
    cacert: |
      -----BEGIN CERTIFICATE-----
      <CA 证书内容>
      -----END CERTIFICATE-----
    cert: |
      -----BEGIN CERTIFICATE-----
      <客户端证书内容>
      -----END CERTIFICATE-----
    key: |
      -----BEGIN RSA PRIVATE KEY-----
      <客户端密钥内容>
      -----END RSA PRIVATE KEY-----
```

## 快速配置脚本

运行以下脚本自动提取并配置：

```bash
#!/bin/bash

# 提取 Kubernetes 认证信息
API_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
echo "API Server: $API_SERVER"

# 提取证书
kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[0].cluster.certificate-authority-data}' | base64 -d > ca.crt
kubectl config view --raw --minify --flatten -o jsonpath='{.users[0].user.client-certificate-data}' | base64 -d > client.crt
kubectl config view --raw --minify --flatten -o jsonpath='{.users[0].user.client-key-data}' | base64 -d > client.key

echo "证书已保存到当前目录："
echo "  - ca.crt"
echo "  - client.crt"
echo "  - client.key"
echo ""
echo "请更新 config.yaml 中的以下配置："
echo "  host: \"$API_SERVER\""
echo "  cacertFile: \"$(pwd)/ca.crt\""
echo "  certFile: \"$(pwd)/client.crt\""
echo "  keyFile: \"$(pwd)/client.key\""
```

## 验证配置

配置完成后，重启 ContainerSSH 并测试连接：

```bash
# 1. 重启服务
./bin/containerssh --config config.yaml

# 2. 在另一个终端测试 SSH 连接
ssh user1@localhost -p 2222
# 输入密码：password1

# 3. 成功后应该能看到 pod 的 shell
```

## 故障排查

### 问题 1：证书验证失败

```
error: x509: certificate signed by unknown authority
```

**解决方案**：确保 `cacertFile` 指向正确的 CA 证书文件。

### 问题 2：权限不足

```
error: pods is forbidden: User "xxx" cannot get resource "pods/exec"
```

**解决方案**：确保使用的 Kubernetes 用户/ServiceAccount 有足够的权限执行 pod exec 操作。

### 问题 3：无法连接到 API Server

```
error: dial tcp: connect: connection refused
```

**解决方案**：
1. 检查 API Server 地址是否正确
2. 确保网络可达
3. 检查防火墙设置

## 安全建议

1. **不要在生产环境中使用 admin 证书**：创建专用的 ServiceAccount 并授予最小权限
2. **保护证书文件**：设置适当的文件权限（600）
3. **定期轮换证书**：建立证书轮换机制
4. **使用 RBAC**：限制 ContainerSSH 只能访问特定 namespace 的特定 pod

## 下一步

配置完成后，你可以：
1. 在 `webhook.yaml` 中为不同用户配置不同的 pod 映射
2. 测试 SSH 连接到 Kubernetes pod
3. 查看日志排查问题
