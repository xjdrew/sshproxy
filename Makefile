.PHONY: build clean run-containerssh run-webhook test

# 构建所有二进制文件
build:
	@echo "Building containerssh..."
	@go build -o bin/containerssh ./cmd/containerssh
	@echo "Building sshhook..."
	@go build -o bin/sshhook ./cmd/sshhook
	@echo "Build complete!"

# 清理构建产物
clean:
	@rm -rf bin/
	@echo "Clean complete!"

# 运行 containerssh
run-containerssh:
	@./bin/containerssh --config config.yaml

# 运行 webhook 服务
run-webhook:
	@./bin/sshhook --config webhook.yaml

# 测试编译
test:
	@go build ./cmd/containerssh
	@go build ./cmd/sshhook
	@echo "Test build successful!"
