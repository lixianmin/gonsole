# Makefile for gonsole project
# 基于项目宪法定义的标准化操作

.PHONY: test web build clean vet fmt

# 默认目标
default: test

# 运行所有测试
test:
	go test -v ./...

# 构建 Web 服务
web:
	go build -o bin/web ./examples/demo.go

# 构建项目
build:
	go build ./...

# 代码格式化
fmt:
	go fmt ./...

# 静态检查
vet:
	go vet ./...

# 清理构建产物
clean:
	rm -rf bin/
