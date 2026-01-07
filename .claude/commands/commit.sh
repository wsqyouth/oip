#!/bin/bash
# .claude/commands/commit.sh
# 智能提交命令 - 支持 rebase、amend 等操作

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 使用说明
usage() {
    echo "用法: commit.sh <message> [选项]"
    echo ""
    echo "选项:"
    echo "  --rebase [branch]   在提交前 rebase 到指定分支（默认: main）"
    echo "  --amend             修改上次提交"
    echo "  --no-verify         跳过代码检查（不推荐）"
    echo "  --push              提交后自动推送（跳过确认）"
    echo ""
    echo "示例:"
    echo "  commit.sh \"feat: 实现订单API\""
    echo "  commit.sh \"feat: 实现订单API\" --rebase"
    echo "  commit.sh \"feat: 实现订单API\" --rebase main"
    echo "  commit.sh \"fix: 修复bug\" --amend"
    exit 1
}

# 参数检查
if [ -z "$1" ]; then
    echo -e "${RED}❌ 错误：缺少提交信息${NC}"
    usage
fi

COMMIT_MSG="$1"
shift

# 默认参数
REBASE=false
REBASE_BRANCH="main"
AMEND=false
NO_VERIFY=false
AUTO_PUSH=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --rebase)
            REBASE=true
            if [[ -n "$2" && "$2" != --* ]]; then
                REBASE_BRANCH="$2"
                shift
            fi
            shift
            ;;
        --amend)
            AMEND=true
            shift
            ;;
        --no-verify)
            NO_VERIFY=true
            shift
            ;;
        --push)
            AUTO_PUSH=true
            shift
            ;;
        *)
            echo -e "${RED}❌ 未知选项: $1${NC}"
            usage
            ;;
    esac
done

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  智能提交流程${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# ============================================
# 1. 代码检查
# ============================================
if [ "$NO_VERIFY" = false ]; then
    echo ""
    echo -e "${YELLOW}🔍 执行代码检查...${NC}"

    # 检查是否是 Go 项目
    IS_GO_PROJECT=false
    if [ -f "go.mod" ] || [ -f "go.work" ]; then
        IS_GO_PROJECT=true
    fi

    if [ "$IS_GO_PROJECT" = true ]; then
        # 1.1 格式化代码
        echo -e "  ${BLUE}→${NC} 格式化代码..."
        gofmt -w . 2>/dev/null || true

        # 1.2 静态检查
        echo -e "  ${BLUE}→${NC} 静态检查..."
        if ! go vet ./... 2>&1 | head -10; then
            echo ""
            echo -e "${YELLOW}⚠️  静态检查发现问题${NC}"
            read -p "是否继续？(y/n) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${RED}❌ 提交已取消${NC}"
                exit 1
            fi
        fi

        # 1.3 运行测试
        echo -e "  ${BLUE}→${NC} 运行单元测试..."
        if ! go test ./... -short -timeout 30s 2>&1 | head -20; then
            echo ""
            echo -e "${YELLOW}⚠️  测试失败${NC}"
            read -p "是否继续？(y/n) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${RED}❌ 提交已取消${NC}"
                exit 1
            fi
        fi

        # 1.4 整理依赖
        echo -e "  ${BLUE}→${NC} 整理依赖..."
        if [ -f "go.work" ]; then
            for dir in */; do
                if [ -f "${dir}go.mod" ]; then
                    (cd "$dir" && go mod tidy 2>/dev/null || true)
                fi
            done
        else
            go mod tidy 2>/dev/null || true
        fi

        echo -e "${GREEN}✅ 代码检查通过${NC}"
    else
        echo -e "${YELLOW}⚠️  非 Go 项目，跳过代码检查${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  跳过代码检查 (--no-verify)${NC}"
fi

# ============================================
# 2. Rebase 操作
# ============================================
if [ "$REBASE" = true ]; then
    echo ""
    echo -e "${YELLOW}🔄 执行 Rebase 操作...${NC}"

    # 2.1 保存当前分支
    CURRENT_BRANCH=$(git branch --show-current)
    echo -e "  ${BLUE}→${NC} 当前分支: ${GREEN}$CURRENT_BRANCH${NC}"
    echo -e "  ${BLUE}→${NC} 目标分支: ${GREEN}$REBASE_BRANCH${NC}"

    # 2.2 检查是否有未提交的改动
    if ! git diff --quiet || ! git diff --cached --quiet; then
        echo -e "  ${BLUE}→${NC} 暂存当前改动..."
        git add .
    fi

    # 2.3 更新目标分支
    echo -e "  ${BLUE}→${NC} 拉取最新 $REBASE_BRANCH..."
    git fetch origin "$REBASE_BRANCH" 2>/dev/null || {
        echo -e "${YELLOW}⚠️  无法拉取远程分支，使用本地分支${NC}"
    }

    # 2.4 执行 rebase
    echo -e "  ${BLUE}→${NC} Rebase 到 $REBASE_BRANCH..."
    if ! git rebase "origin/$REBASE_BRANCH" 2>/dev/null && ! git rebase "$REBASE_BRANCH" 2>/dev/null; then
        echo ""
        echo -e "${RED}❌ Rebase 失败！检测到冲突${NC}"
        echo ""
        echo -e "${YELLOW}请手动解决冲突后执行：${NC}"
        echo -e "  1. 解决冲突"
        echo -e "  2. git add <解决的文件>"
        echo -e "  3. git rebase --continue"
        echo -e "  4. 重新运行此脚本"
        echo ""
        echo -e "${BLUE}或者取消 rebase：${NC}"
        echo -e "  git rebase --abort"
        exit 1
    fi

    echo -e "${GREEN}✅ Rebase 成功${NC}"
fi

# ============================================
# 3. Git 提交
# ============================================
echo ""
echo -e "${YELLOW}📊 Git 状态:${NC}"
git status -s

echo ""
if [ "$AMEND" = true ]; then
    echo -e "${YELLOW}📝 修改上次提交...${NC}"
    read -p "确认修改上次提交？(y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}❌ 提交已取消${NC}"
        exit 1
    fi

    git add .
    git commit --amend -m "$COMMIT_MSG

🤖 Generated with Claude Code

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"

else
    echo -e "${YELLOW}📝 提交代码...${NC}"
    read -p "确认提交以上文件？(y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}❌ 提交已取消${NC}"
        exit 1
    fi

    git add .

    # 检查是否有文件被添加
    if git diff --cached --quiet; then
        echo -e "${YELLOW}⚠️  没有文件需要提交${NC}"
        exit 0
    fi

    git commit -m "$COMMIT_MSG

🤖 Generated with Claude Code

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
fi

echo -e "${GREEN}✅ 代码已提交: $COMMIT_MSG${NC}"

# ============================================
# 4. 推送代码
# ============================================
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}📌 下一步操作：${NC}"
echo -e "   推送到远程: ${GREEN}git push${NC}"
if [ "$REBASE" = true ]; then
    echo -e "   强制推送: ${GREEN}git push --force-with-lease${NC} ${YELLOW}(推荐)${NC}"
fi
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# 检查当前分支是否有 upstream
CURRENT_BRANCH=$(git branch --show-current)
HAS_UPSTREAM=$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || echo "")

if [ "$AUTO_PUSH" = true ]; then
    echo ""
    echo -e "${YELLOW}🚀 自动推送模式...${NC}"

    if [ -z "$HAS_UPSTREAM" ]; then
        # 首次推送，需要设置 upstream
        echo -e "${YELLOW}  → 首次推送，设置 upstream...${NC}"
        git push -u origin "$CURRENT_BRANCH"
    elif [ "$REBASE" = true ]; then
        git push --force-with-lease
    else
        git push
    fi
    echo -e "${GREEN}✅ 代码已推送到远程${NC}"
else
    echo ""
    read -p "是否现在推送到远程？(y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [ -z "$HAS_UPSTREAM" ]; then
            # 首次推送，需要设置 upstream
            echo -e "${YELLOW}  → 首次推送，设置 upstream...${NC}"
            git push -u origin "$CURRENT_BRANCH"
            echo -e "${GREEN}✅ 代码已推送到远程${NC}"
        elif [ "$REBASE" = true ]; then
            echo -e "${YELLOW}⚠️  检测到 rebase，使用 --force-with-lease 推送${NC}"
            read -p "确认强制推送？(y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                git push --force-with-lease
                echo -e "${GREEN}✅ 代码已推送到远程${NC}"
            else
                echo -e "${YELLOW}⏸️  推送已取消${NC}"
            fi
        else
            git push
            echo -e "${GREEN}✅ 代码已推送到远程${NC}"
        fi
    else
        echo -e "${YELLOW}⏸️  推送已跳过，稍后可手动执行:${NC}"
        if [ -z "$HAS_UPSTREAM" ]; then
            echo -e "   ${GREEN}git push -u origin $CURRENT_BRANCH${NC}"
        elif [ "$REBASE" = true ]; then
            echo -e "   ${GREEN}git push --force-with-lease${NC}"
        else
            echo -e "   ${GREEN}git push${NC}"
        fi
    fi
fi

# ============================================
# 5. 更新提交日志
# ============================================
mkdir -p tasks
echo "✅ 已提交: $COMMIT_MSG ($(date +%Y-%m-%d\ %H:%M))" >> tasks/commit-log.md

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  ✨ 提交流程完成！${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
