#!/bin/bash

set -e

# 全局变量
DPMAIN_PID=""
DPSYNC_PID=""
LOG_DIR="/tmp/oip-verify-logs"
DPMAIN_LOG="$LOG_DIR/dpmain.log"
DPSYNC_LOG="$LOG_DIR/dpsync.log"

# 清理函数
cleanup() {
    echo ""
    echo "========================================="
    echo "  清理测试环境..."
    echo "========================================="

    if [ -n "$DPMAIN_PID" ] && kill -0 "$DPMAIN_PID" 2>/dev/null; then
        echo "   -> 停止 dpmain 服务 (PID: $DPMAIN_PID)..."
        kill -TERM "$DPMAIN_PID" 2>/dev/null || true
        sleep 2
        kill -0 "$DPMAIN_PID" 2>/dev/null && kill -9 "$DPMAIN_PID" 2>/dev/null || true
        echo "   ✓ dpmain 服务已停止"
    fi

    if [ -n "$DPSYNC_PID" ] && kill -0 "$DPSYNC_PID" 2>/dev/null; then
        echo "   -> 停止 dpsync 服务 (PID: $DPSYNC_PID)..."
        kill -TERM "$DPSYNC_PID" 2>/dev/null || true
        sleep 2
        kill -0 "$DPSYNC_PID" 2>/dev/null && kill -9 "$DPSYNC_PID" 2>/dev/null || true
        echo "   ✓ dpsync 服务已停止"
    fi

    echo "   ✓ 清理完成"
    echo ""
    echo "日志文件位置："
    echo "   - dpmain: $DPMAIN_LOG"
    echo "   - dpsync: $DPSYNC_LOG"
    echo ""
}

# 注册清理函数
trap cleanup EXIT INT TERM

echo "========================================="
echo "  OIP Backend - 验证构建脚本"
echo "========================================="
echo ""

# 创建日志目录
mkdir -p "$LOG_DIR"

# 检查 Go 版本
echo "1. 检查 Go 版本..."
go version
echo ""

# 检查当前目录
echo "2. 当前目录: $(pwd)"
echo ""

# 清理可能的缓存问题
echo "3. 清理模块缓存..."
#go clean -modcache 2>/dev/null || echo "   (跳过缓存清理)"
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

# 第8步：启动服务
echo "8. 启动测试服务..."
echo ""

# 检查端口占用
echo "   -> 检查端口占用..."
if lsof -i :8080 >/dev/null 2>&1; then
    echo "   ✗ 端口 8080 已被占用，请先停止占用该端口的进程"
    lsof -i :8080 | grep LISTEN
    exit 1
fi
echo "   ✓ 端口 8080 可用"

# 检查 Docker 服务
echo "   -> 检查 Docker 服务..."
if ! docker ps >/dev/null 2>&1; then
    echo "   ✗ Docker 未运行，请先启动 Docker"
    exit 1
fi

DOCKER_SERVICES=$(docker ps --format '{{.Names}}' | grep -E 'oip_(mysql|redis|lmstfy)' | wc -l | tr -d ' ')
if [ "$DOCKER_SERVICES" -lt 3 ]; then
    echo "   ⚠️  警告: Docker 服务未完全启动 (当前: $DOCKER_SERVICES/3)"
    echo "   尝试启动 Docker 服务..."
    docker-compose up -d 2>&1 | grep -v "WARNING"
    sleep 3
fi
echo "   ✓ Docker 服务运行正常 (MySQL, Redis, Lmstfy)"

# 启动 dpmain 服务
echo "   -> 启动 dpmain 服务..."
cd dpmain
if [ ! -f "bin/dpmain-apiserver" ]; then
    echo "   ✗ dpmain 二进制文件不存在，构建应该在步骤5完成"
    exit 1
fi

./bin/dpmain-apiserver > "$DPMAIN_LOG" 2>&1 &
DPMAIN_PID=$!
echo "   ✓ dpmain 已启动 (PID: $DPMAIN_PID)"

# 等待 dpmain 启动
echo "   -> 等待 dpmain 服务就绪..."
MAX_WAIT=30
WAIT_COUNT=0
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if curl -s http://localhost:8080/health >/dev/null 2>&1; then
        echo "   ✓ dpmain 服务就绪 (耗时: ${WAIT_COUNT}s)"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    printf "."
done
echo ""

if [ $WAIT_COUNT -ge $MAX_WAIT ]; then
    echo "   ✗ dpmain 服务启动超时"
    echo "   日志文件: $DPMAIN_LOG"
    echo "   最近日志:"
    tail -20 "$DPMAIN_LOG"
    exit 1
fi

# 启动 dpsync 服务
cd ../dpsync
echo "   -> 启动 dpsync 服务..."
if [ ! -f "bin/dpsync-worker" ]; then
    echo "   ✗ dpsync 二进制文件不存在，构建应该在步骤6完成"
    exit 1
fi

./bin/dpsync-worker > "$DPSYNC_LOG" 2>&1 &
DPSYNC_PID=$!
echo "   ✓ dpsync 已启动 (PID: $DPSYNC_PID)"

# 等待 dpsync 启动
echo "   -> 等待 dpsync 服务就绪..."
sleep 3

# 检查 dpsync 进程是否还在运行
if ! kill -0 "$DPSYNC_PID" 2>/dev/null; then
    echo "   ✗ dpsync 服务启动失败"
    echo "   日志文件: $DPSYNC_LOG"
    echo "   最近日志:"
    tail -20 "$DPSYNC_LOG"
    exit 1
fi
echo "   ✓ dpsync 服务就绪"

cd ..
echo ""
echo "   ✓✓✓ 所有服务启动成功！"
echo "       - dpmain:  http://localhost:8080 (PID: $DPMAIN_PID)"
echo "       - dpsync:  Worker 运行中 (PID: $DPSYNC_PID)"
echo ""
echo "   日志位置:"
echo "       - dpmain: $DPMAIN_LOG"
echo "       - dpsync: $DPSYNC_LOG"
echo ""

# 第9步：E2E 测试
echo "9. E2E 测试：订单创建与诊断完整链路..."
echo ""

# 检查 jq 是否安装
if ! command -v jq &> /dev/null; then
    echo "   ✗ jq 未安装，无法执行 E2E 测试"
    echo "   安装方法: brew install jq (macOS) 或 apt-get install jq (Linux)"
    exit 1
fi

# 创建测试账户
echo "   -> 步骤 1/4: 创建测试账户..."
TIMESTAMP=$(date +%s)
ACCOUNT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/accounts \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"E2E Test Account\",\"email\":\"e2e-test-$TIMESTAMP@example.com\"}")

ACCOUNT_CODE=$(echo "$ACCOUNT_RESPONSE" | jq -r '.code // .meta.code // 0')
if [ "$ACCOUNT_CODE" != "200" ]; then
    echo "   ✗ 创建账户失败"
    echo "   响应: $ACCOUNT_RESPONSE"
    echo ""
    echo "   检查 dpmain 日志:"
    tail -20 "$DPMAIN_LOG"
    exit 1
fi

ACCOUNT_ID=$(echo "$ACCOUNT_RESPONSE" | jq -r '.data.id')
echo "   ✓ 账户创建成功 (ID: $ACCOUNT_ID)"

# 验证账户可查询
echo "   -> 步骤 2/4: 验证账户查询..."
ACCOUNT_GET=$(curl -s "http://localhost:8080/api/v1/accounts/$ACCOUNT_ID")
ACCOUNT_GET_CODE=$(echo "$ACCOUNT_GET" | jq -r '.code // .meta.code // 0')
if [ "$ACCOUNT_GET_CODE" != "200" ]; then
    echo "   ✗ 账户查询失败"
    exit 1
fi
echo "   ✓ 账户查询成功"

# 创建订单（Smart Wait 模式）
echo "   -> 步骤 3/4: 创建订单并等待诊断 (Smart Wait 15秒)..."
ORDER_NO="E2E-TEST-$TIMESTAMP"

# 定义订单测试数据（可根据需要修改）
ORDER_PAYLOAD=$(cat <<EOF
{
    "account_id": $ACCOUNT_ID,
    "merchant_order_no": "$ORDER_NO",
    "shipment": {
        "ship_from": {
            "contact_name": "E2E Test Store",
            "street1": "230 W 200 S",
            "city": "Salt Lake City",
            "state": "UT",
            "postal_code": "84101",
            "country": "US",
            "phone": "+1-801-555-0100",
            "email": "store@e2etest.com"
        },
        "ship_to": {
            "contact_name": "E2E Test Customer",
            "street1": "123 Main St",
            "city": "Seattle",
            "state": "WA",
            "postal_code": "98101",
            "country": "US",
            "phone": "+1-206-555-0200",
            "email": "customer@e2etest.com"
        },
        "parcels": [
            {
                "weight": {"value": 1.5, "unit": "kg"},
                "dimension": {"width": 20, "height": 15, "depth": 10, "unit": "cm"},
                "items": [
                    {
                        "description": "E2E Test Item",
                        "quantity": 2,
                        "price": {"amount": 19.99, "currency": "USD"},
                        "sku": "E2E-TEST-001"
                    }
                ]
            }
        ]
    }
}
EOF
)

ORDER_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/orders?wait=15" \
    -H "Content-Type: application/json" \
    -d "$ORDER_PAYLOAD")

ORDER_CODE=$(echo "$ORDER_RESPONSE" | jq -r '.code // .meta.code // 0')
ORDER_ID=$(echo "$ORDER_RESPONSE" | jq -r '.data.id // .data.order_id // empty')

if [ "$ORDER_CODE" = "200" ]; then
    # 诊断在 15 秒内完成
    ORDER_STATUS=$(echo "$ORDER_RESPONSE" | jq -r '.data.status // empty')
    echo "   ✓ 订单创建成功 (ID: $ORDER_ID)"
    echo "   ✓ 诊断已完成 (状态: $ORDER_STATUS)"

    # 验证诊断结果
    DIAGNOSIS_ITEMS=$(echo "$ORDER_RESPONSE" | jq -r '.data.diagnosis.items // [] | length')
    if [ "$DIAGNOSIS_ITEMS" -gt 0 ]; then
        echo "   ✓ 诊断结果包含 $DIAGNOSIS_ITEMS 个项目"

        # 显示诊断项目详情
        DIAGNOSIS_TYPES=$(echo "$ORDER_RESPONSE" | jq -r '.data.diagnosis.items[].type' | tr '\n' ', ' | sed 's/,$//')
        echo "   ✓ 诊断类型: [$DIAGNOSIS_TYPES]"

        # 验证订单查询
        echo "   -> 步骤 4/4: 验证订单查询..."
        ORDER_GET=$(curl -s "http://localhost:8080/api/v1/orders/$ORDER_ID")
        ORDER_GET_CODE=$(echo "$ORDER_GET" | jq -r '.code // .meta.code // 0')
        if [ "$ORDER_GET_CODE" != "200" ]; then
            echo "   ✗ 订单查询失败"
            exit 1
        fi
        echo "   ✓ 订单查询成功"

        echo ""
        echo "   ✓✓✓ E2E 测试完整链路验证成功！"
        echo ""
        echo "   验证的完整链路："
        echo "       1. [dpmain] 接收订单创建请求 ✓"
        echo "       2. [dpmain] 保存订单到 MySQL ✓"
        echo "       3. [dpmain] 推送诊断任务到 Lmstfy 队列 (order_diagnose) ✓"
        echo "       4. [dpsync] 从队列消费诊断任务 ✓"
        echo "       5. [dpsync] 执行诊断逻辑 (shipping, anomaly) ✓"
        echo "       6. [dpsync] 推送回调到 callback 队列 (order_diagnose_callback) ✓"
        echo "       7. [dpmain] Callback Consumer 消费回调 ✓"
        echo "       8. [dpmain] 更新订单状态和诊断结果到 MySQL ✓"
        echo "       9. [dpmain] 通过 Redis Pub/Sub 通知等待的 API 请求 ✓"
        echo "      10. [dpmain] API 返回完整诊断结果给客户端 ✓"
        echo ""
        echo "   测试数据："
        echo "       - 账户ID: $ACCOUNT_ID"
        echo "       - 订单ID: $ORDER_ID"
        echo "       - 订单号: $ORDER_NO"
        echo "       - 诊断项: $DIAGNOSIS_ITEMS 个 ($DIAGNOSIS_TYPES)"
    else
        echo "   ✗ 诊断结果为空"
        echo "   响应: $ORDER_RESPONSE"
        exit 1
    fi

elif [ "$ORDER_CODE" = "3001" ]; then
    # 诊断超时，仍在处理中
    echo "   ⚠️  订单创建成功但诊断超时 (ID: $ORDER_ID, Code: 3001)"
    echo ""
    echo "   这不是预期结果。检查项："
    echo "       - dpsync 服务是否正常运行？"
    echo "       - callback consumer 是否正常消费？"
    echo "       - Lmstfy 队列是否正常？"
    echo ""
    echo "   检查 dpsync 日志:"
    tail -30 "$DPSYNC_LOG"
    echo ""
    echo "   检查 dpmain 日志:"
    tail -30 "$DPMAIN_LOG"
    exit 1

else
    echo "   ✗ 订单创建失败 (Code: $ORDER_CODE)"
    echo "   响应: $ORDER_RESPONSE"
    echo ""
    echo "   检查 dpmain 日志:"
    tail -30 "$DPMAIN_LOG"
    exit 1
fi

echo ""
echo "========================================="
echo "  ✓✓✓ 所有测试通过！"
echo "========================================="
echo ""
echo "日志文件位置："
echo "  - dpmain: $DPMAIN_LOG"
echo "  - dpsync: $DPSYNC_LOG"
echo ""
echo "提示: 服务将在脚本退出时自动停止"
echo ""
