#!/bin/bash
# .claude/commands/commit-push.sh
# 一键提交并推送代码

set -e

# 参数检查
if [ -z "$1" ]; then
    echo "❌ 错误：缺少提交信息"
    echo "用法: ./.claude/commands/commit-push.sh \"commit message\""
    echo ""
    echo "示例:"
    echo "  ./.claude/commands/commit-push.sh \"feat: 实现订单创建API\""
    echo "  ./.claude/commands/commit-push.sh \"fix: 修复Redis连接泄漏\""
    exit 1
fi

COMMIT_MSG="$1"

echo "🔍 执行代码检查..."

# 检查是否是 Go 项目
IS_GO_PROJECT=false
if [ -f "go.mod" ] || [ -f "go.work" ]; then
    IS_GO_PROJECT=true
fi

if [ "$IS_GO_PROJECT" = true ]; then
    # 1. 格式化代码
    echo "  - 格式化代码..."
    gofmt -w .

    # 2. 静态检查
    echo "  - 静态检查..."
    if ! go vet ./... 2>&1 | head -20; then
        echo ""
        echo "⚠️  静态检查发现问题"
        read -p "是否继续提交？(y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ 提交已取消"
            exit 1
        fi
    fi

    # 3. 运行测试（跳过集成测试）
    echo "  - 运行单元测试..."
    if ! go test ./... -short -timeout 30s 2>&1 | head -50; then
        echo ""
        echo "⚠️  测试失败"
        read -p "是否继续提交？(y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "❌ 提交已取消"
            exit 1
        fi
    fi

    # 4. 整理依赖
    echo "  - 整理依赖..."
    if [ -f "go.work" ]; then
        # Go Workspace: 整理所有模块
        for dir in */; do
            if [ -f "${dir}go.mod" ]; then
                echo "    - 整理 ${dir}..."
                (cd "$dir" && go mod tidy)
            fi
        done
    else
        go mod tidy
    fi

    echo "✅ 代码检查通过"
else
    echo "⚠️  非 Go 项目，跳过代码检查"
fi

# 5. 检查 Git 状态
echo ""
echo "📊 Git 状态:"
git status -s

echo ""
read -p "确认提交以上文件？(y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ 提交已取消"
    exit 1
fi

# 6. Git 操作
echo ""
echo "📝 提交代码..."
git add .

# 检查是否有文件被添加
if git diff --cached --quiet; then
    echo "⚠️  没有文件需要提交"
    exit 0
fi

git commit -m "$COMMIT_MSG

🤖 Generated with Claude Code

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"

echo ""
echo "✅ 代码已提交: $COMMIT_MSG"
echo ""
echo "📌 下一步操作："
echo "   推送到远程: git push"
echo "   或强制推送: git push --force-with-lease"
echo ""
read -p "是否现在推送到远程？(y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🚀 推送代码..."
    git push
    echo "✅ 代码已推送到远程"
else
    echo "⏸️  推送已跳过，稍后可手动执行: git push"
fi

# 7. 更新提交日志
mkdir -p tasks
echo "✅ 已提交: $COMMIT_MSG ($(date +%Y-%m-%d\ %H:%M))" >> tasks/commit-log.md

echo ""
echo "📋 提交日志已更新: tasks/commit-log.md"
