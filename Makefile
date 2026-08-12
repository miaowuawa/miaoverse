# 跨平台构建 / 常用开发命令入口
# Windows 用户直接使用 PowerShell: .\scripts\build.ps1

BACKEND_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GO := go

.PHONY: all build build-linux build-windows run test fmt clean

all: build

# 构建当前平台
build:
	mkdir -p $(BACKEND_DIR)/dist && cd $(BACKEND_DIR) && CGO_ENABLED=0 $(GO) build -trimpath -o dist/miaoverse .

# 构建 linux amd64 版本
build-linux:
	mkdir -p $(BACKEND_DIR)/dist && cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -o dist/miaoverse-linux-amd64 .

# 构建 windows amd64 版本
build-windows:
	mkdir -p $(BACKEND_DIR)/dist && cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -o dist/miaoverse-windows-amd64.exe .

run:
	cd $(BACKEND_DIR) && $(GO) run .

test:
	cd $(BACKEND_DIR) && $(GO) test ./...

fmt:
	cd $(BACKEND_DIR) && gofmt -w .

clean:
	rm -rf $(BACKEND_DIR)/dist
