#!/bin/bash
# .claude/hooks/post-tool-use.sh
# 代码完成后自动执行检查

# 注意：这个脚本由 Claude Code 在编辑文件后自动触发
# 不需要手动运行

set -e

echo "🔍 执行代码检查..."

# 1. 检查是否在项目目录
if [ ! -f "go.mod" ] && [ ! -f "go.work" ]; then
    echo "⚠️  不在 Go 项目目录，跳过检查"
    exit 0
fi

# 2. 格式化检查
echo "  - 检查代码格式..."
UNFORMATTED=$(gofmt -l . | grep -v vendor || true)
if [ -n "$UNFORMATTED" ]; then
    echo "⚠️  以下文件需要格式化："
    echo "$UNFORMATTED"
    echo "  自动格式化中..."
    gofmt -w .
    echo "✅ 代码已格式化"
else
    echo "✅ 代码格式正确"
fi

# 3. 检查冗余导入
echo "  - 检查冗余导入..."
if [ -f "go.work" ]; then
    # Go Workspace: 整理所有模块
    for dir in */; do
        if [ -f "${dir}go.mod" ]; then
            (cd "$dir" && go mod tidy 2>/dev/null || true)
        fi
    done
else
    go mod tidy 2>/dev/null || true
fi

# 4. 静态分析（仅警告）
echo "  - 静态分析..."
go vet ./... 2>&1 | head -20 || {
    echo "⚠️  静态检查发现问题，请查看上述输出"
}

# 5. 检查文档同步（TODO: 实现文档同步检查器）
# python3 .claude/agents/doc-sync-checker.py

echo "✅ 代码检查完成"
