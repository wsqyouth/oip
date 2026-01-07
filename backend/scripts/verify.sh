#!/bin/bash

set -e

echo "========================================="
echo "  OIP Backend - 验证构建脚本"
echo "========================================="
echo ""

# 检查 Go 版本
echo "1. 检查 Go 版本..."
go version
echo ""

# 检查当前目录
echo "2. 当前目录: $(pwd)"
echo ""

# 清理可能的缓存问题
echo "3. 清理模块缓存..."
go clean -modcache 2>/dev/null || echo "   (跳过缓存清理)"
echo ""

# 验证 common 模块
echo "4. 验证 common 模块..."
cd common
echo "   -> go mod tidy"
go mod tidy
echo "   ✓ common 模块验证成功"
cd ..
echo ""

# 验证 dpmain 模块
echo "5. 验证 dpmain 模块..."
cd dpmain
echo "   -> go mod tidy"
go mod tidy
echo "   -> go build"
go build -o bin/dpmain-apiserver ./cmd/apiserver
echo "   ✓ dpmain 模块构建成功: $(ls -lh bin/dpmain-apiserver | awk '{print $5, $9}')"
cd ..
echo ""

# 验证 dpsync 模块
echo "6. 验证 dpsync 模块..."
cd dpsync
echo "   -> go mod tidy"
go mod tidy
echo "   -> go build"
go build -o bin/dpsync-worker ./cmd/worker
echo "   ✓ dpsync 模块构建成功: $(ls -lh bin/dpsync-worker | awk '{print $5, $9}')"
cd ..
echo ""

# 验证 Go Workspace
echo "7. 验证 Go Workspace..."
go work sync
echo "   ✓ Workspace 同步成功"
echo ""

echo "========================================="
echo "  ✓ 所有模块验证通过！"
echo "========================================="
echo ""
echo "下一步："
echo "  make run-dpmain    # 启动 API 服务"
echo "  make run-dpsync    # 启动 Worker 服务"
echo ""
